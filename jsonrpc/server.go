// Copyright (c) 2013-2017 The btcsuite developers
// Copyright (c) 2015-2017 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package jsonrpc

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/uxuycom/indexer/storage"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the
	// RPC server is allowed to stay open without authenticating before it
	// is closed.
	rpcAuthTimeoutSeconds = 10
)

var (
	// JSON 2.0 batched request prefix
	batchedRequestPrefix = []byte("[")

	rpcsLog = logrus.StandardLogger()

	jsonParamIdReg = regexp.MustCompile(`"id":\s*(\d+)`)
)

var (
	cfg = &Config{}
)

// Config defines the configuration options for btcd.
//
// See loadConfig for details on the configuration load process.
type Config struct {
	DebugLevel           string   `json:"debug_level"`
	DisableTLS           bool     `json:"notls" description:"Disable TLS for the RPC server -- NOTE: This is only allowed if the RPC server is bound to localhost"`
	RPCCert              string   `json:"rpccert" description:"File containing the certificate file"`
	RPCKey               string   `json:"rpckey" description:"File containing the certificate key"`
	RPCLimitPass         string   `json:"rpclimitpass" default-mask:"-" description:"Password for limited RPC connections"`
	RPCLimitUser         string   `json:"rpclimituser" description:"Username for limited RPC connections"`
	RPCListeners         []string `json:"rpclisten" description:"Add an interface/port to listen for RPC connections (default port: 6583, testnet: 16583)"`
	RPCMaxClients        int      `json:"rpcmaxclients" description:"Max number of RPC clients for standard connections"`
	RPCMaxConcurrentReqs int      `json:"rpcmaxconcurrentreqs" description:"Max number of concurrent RPC requests that may be processed concurrently"`
	RPCMaxWebsockets     int      `json:"rpcmaxwebsockets" description:"Max number of RPC websocket connections"`
	RPCQuirks            bool     `json:"rpcquirks" description:"Mirror some JSON-RPC quirks of Bitcoin Core -- NOTE: Discouraged unless interoperability issues need to be worked around"`
	RPCPass              string   `json:"rpcpass" default-mask:"-" description:"Password for RPC connections"`
	RPCUser              string   `json:"rpcuser" description:"Username for RPC connections"`
}

type commandHandler func(*RpcServer, interface{}, <-chan struct{}) (interface{}, error)

// rpcHandlers maps RPC command strings to appropriate handler functions.
// This is set by init because help references rpcHandlers and thus causes
// a dependency loop.
var rpcHandlers map[string]commandHandler

// Commands that are available to a limited user
var rpcLimited = map[string]struct{}{}

// internalRPCError is a convenience function to convert an internal error to
// an RPC error with the appropriate code set.  It also logs the error to the
// RPC server subsystem since internal xyerrors really should not occur.  The
// context parameter is only used in the xylog message and may be empty if it's
// not needed.
func internalRPCError(errStr, context string) *RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	rpcsLog.Error(logStr)
	return NewRPCError(ErrRPCInternal.Code, errStr)
}

// RpcServer provides a concurrent safe RPC server to a chain server.
type RpcServer struct {
	started                int32
	shutdown               int32
	cfg                    RpcServerConfig
	authsha                [sha256.Size]byte
	limitauthsha           [sha256.Size]byte
	numClients             int32
	statusLines            map[int]string
	statusLock             sync.RWMutex
	wg                     sync.WaitGroup
	requestProcessShutdown chan struct{}
	quit                   chan int
	dbc                    *storage.DBClient
}

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1)
// for the given request and response status code.  This function was lifted and
// adapted from the standard library HTTP server code since it's not exported.
func (s *RpcServer) httpStatusLine(req *http.Request, code int) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	s.statusLock.RLock()
	line, ok := s.statusLines[key]
	s.statusLock.RUnlock()
	if ok {
		return line
	}

	// Slow path:
	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		line = proto + " " + codeStr + " " + text + "\r\n"
		s.statusLock.Lock()
		s.statusLines[key] = line
		s.statusLock.Unlock()
	} else {
		text = "status code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}

	return line
}

// writeHTTPResponseHeaders writes the necessary response headers prior to
// writing an HTTP body given a request to use for protocol negotiation, headers
// to write, a status code, and a writer.
func (s *RpcServer) writeHTTPResponseHeaders(req *http.Request, headers http.Header, code int, w io.Writer) error {
	_, err := io.WriteString(w, s.httpStatusLine(req, code))
	if err != nil {
		return err
	}

	err = headers.Write(w)
	if err != nil {
		return err
	}

	_, err = io.WriteString(w, "\r\n")
	return err
}

// Stop is used by server.go to stop the rpc listener.
func (s *RpcServer) Stop() error {
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		rpcsLog.Infof("RPC server is already in the process of shutting down")
		return nil
	}
	rpcsLog.Warnf("RPC server shutting down")
	for _, listener := range s.cfg.Listeners {
		err := listener.Close()
		if err != nil {
			rpcsLog.Errorf("Problem shutting down rpc: %v", err)
			return err
		}
	}
	close(s.quit)
	s.wg.Wait()
	rpcsLog.Infof("RPC server shutdown complete")
	return nil
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized
// RPC client requests the process to shutdown.  If the request can not be read
// immediately, it is dropped.
func (s *RpcServer) RequestedProcessShutdown() <-chan struct{} {
	return s.requestProcessShutdown
}

// limitConnections responds with a 503 service unavailable and returns true if
// adding another client would exceed the maximum allow RPC clients.
//
// This function is safe for concurrent access.
func (s *RpcServer) limitConnections(w http.ResponseWriter, remoteAddr string) bool {
	if int(atomic.LoadInt32(&s.numClients)+1) > cfg.RPCMaxClients {
		rpcsLog.Infof("Max RPC clients exceeded [%d] - "+
			"disconnecting client %s", cfg.RPCMaxClients,
			remoteAddr)
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// incrementClients adds one to the number of connected RPC clients.  Note
// this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *RpcServer) incrementClients() {
	atomic.AddInt32(&s.numClients, 1)
}

// decrementClients subtracts one from the number of connected RPC clients.
// Note this only applies to standard clients.  Websocket clients have their own
// limits and are tracked separately.
//
// This function is safe for concurrent access.
func (s *RpcServer) decrementClients() {
	atomic.AddInt32(&s.numClients, -1)
}

// parsedRPCCmd represents a JSON-RPC request object that has been parsed into
// a known concrete command along with any error that might have happened while
// parsing it.
type parsedRPCCmd struct {
	jsonrpc RPCVersion
	id      interface{}
	method  string
	cmd     interface{}
	err     *RPCError
}

// standardCmdResult checks that a parsed command is a standard Bitcoin JSON-RPC
// command and runs the appropriate handler to reply to the command.  Any
// commands which are not recognized or not implemented will return an error
// suitable for use in replies.
func (s *RpcServer) standardCmdResult(cmd *parsedRPCCmd, closeChan <-chan struct{}) (interface{}, error) {
	handler, ok := rpcHandlers[cmd.method]
	if ok {
		return handler(s, cmd.cmd, closeChan)
	}
	return nil, ErrRPCMethodNotFound
}

// parseCmd parses a JSON-RPC request object into known concrete command.  The
// err field of the returned parsedRPCCmd struct will contain an RPC error that
// is suitable for use in replies if the command is invalid in some way such as
// an unregistered command or invalid parameters.
func parseCmd(request *Request) *parsedRPCCmd {
	parsedCmd := parsedRPCCmd{
		jsonrpc: request.Jsonrpc,
		id:      request.ID,
		method:  request.Method,
	}

	cmd, err := UnmarshalCmd(request)
	if err != nil {
		// When the error is because the method is not registered,
		// produce a method not found RPC error.
		if jerr, ok := err.(Error); ok &&
			jerr.ErrorCode == ErrUnregisteredMethod {

			parsedCmd.err = ErrRPCMethodNotFound
			return &parsedCmd
		}

		// Otherwise, some type of invalid parameters is the
		// cause, so produce the equivalent RPC error.
		parsedCmd.err = NewRPCError(
			ErrRPCInvalidParams.Code, err.Error())
		return &parsedCmd
	}

	parsedCmd.cmd = cmd
	return &parsedCmd
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the
// passed parameters.  It will automatically convert xyerrors that are not of
// the type *RPCError to the appropriate type as needed.
func createMarshalledReply(rpcVersion RPCVersion, id interface{}, result interface{}, replyErr error) ([]byte, error) {
	var jsonErr *RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = internalRPCError(replyErr.Error(), "")
		}
	}

	return MarshalResponse(rpcVersion, id, result, jsonErr)
}

// processRequest determines the incoming request type (single or batched),
// parses it and returns a marshalled response.
func (s *RpcServer) processRequest(request *Request, isAdmin bool, closeChan <-chan struct{}) []byte {
	var result interface{}
	var err error
	var jsonErr *RPCError
	if !isAdmin {
		if _, ok := rpcLimited[request.Method]; !ok {
			jsonErr = internalRPCError("limited user not "+
				"authorized for this method", "")
		}
	}

	if jsonErr == nil {
		if request.Method == "" || request.Params == nil {
			jsonErr = &RPCError{
				Code:    ErrRPCInvalidRequest.Code,
				Message: "Invalid request: malformed",
			}
			msg, err := createMarshalledReply(request.Jsonrpc, request.ID, result, jsonErr)
			if err != nil {
				rpcsLog.Errorf("Failed to marshal reply: %v", err)
				return nil
			}
			return msg
		}

		// Valid requests with no ID (notifications) must not have a response
		// per the JSON-RPC spec.
		if request.ID == nil {
			return nil
		}

		// Attempt to parse the JSON-RPC request into a known
		// concrete command.
		parsedCmd := parseCmd(request)
		if parsedCmd.err != nil {
			jsonErr = parsedCmd.err
		} else {
			result, err = s.standardCmdResult(parsedCmd,
				closeChan)
			if err != nil {
				if rpcErr, ok := err.(*RPCError); ok {
					jsonErr = rpcErr
				} else {
					jsonErr = &RPCError{
						Code:    ErrRPCInternal.Code,
						Message: err.Error(),
					}
				}
			}
		}
	}

	// Marshal the response.
	msg, err := createMarshalledReply(request.Jsonrpc, request.ID, result, jsonErr)
	if err != nil {
		rpcsLog.Errorf("Failed to marshal reply: %v", err)
		return nil
	}
	return msg
}

// jsonRPCRead handles reading and responding to RPC messages.
func (s *RpcServer) jsonRPCRead(w http.ResponseWriter, r *http.Request, isAdmin bool) {
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}

	// Read and close the JSON-RPC request body from the caller.
	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %v",
			errCode, err), errCode)
		return
	}

	start := time.Now()

	//match id value
	rid := ""
	matches := jsonParamIdReg.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		rid = matches[1]
	} else {
		rid = uuid.New().String()
	}
	rpcsLog.Infof("rid[%s], start process request, params[%s]", rid, string(body))
	defer func() {
		rpcsLog.Infof("rid[%s], finish process request, cost[%v]", rid, time.Since(start))
	}()

	// Unfortunately, the http server doesn't provide the ability to
	// change the read deadline for the new connection and having one breaks
	// long polling.  However, not having a read deadline on the initial
	// connection would mean clients can connect and idle forever.  Thus,
	// hijack the connecton from the HTTP server, clear the read deadline,
	// and handle writing the response manually.
	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		rpcsLog.Warnf(errMsg)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		rpcsLog.Warnf("Failed to hijack HTTP connection: %v", err)
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer conn.Close()
	defer buf.Flush()

	var timeZeroVal time.Time

	_ = conn.SetReadDeadline(timeZeroVal)

	// Attempt to parse the raw body into a JSON-RPC request.
	// Setup a close notifier.  Since the connection is hijacked,
	// the CloseNotifer on the ResponseWriter is not available.
	closeChan := make(chan struct{}, 1)
	go func() {
		_, err = conn.Read(make([]byte, 1))
		if err != nil {
			close(closeChan)
		}
	}()

	var results []json.RawMessage
	var batchSize int
	var batchedRequest bool

	// Determine request type
	if bytes.HasPrefix(body, batchedRequestPrefix) {
		batchedRequest = true
	}

	// Process a single request
	if !batchedRequest {
		var req Request
		var resp json.RawMessage
		err = json.Unmarshal(body, &req)
		if err != nil {
			jsonErr := &RPCError{
				Code: ErrRPCParse.Code,
				Message: fmt.Sprintf("Failed to parse request: %v",
					err),
			}
			resp, err = MarshalResponse(RpcVersion1, nil, nil, jsonErr)
			if err != nil {
				rpcsLog.Errorf("Failed to create reply: %v", err)
			}
		}

		if err == nil {
			// The JSON-RPC 1.0 spec defines that notifications must have their "id"
			// set to null and states that notifications do not have a response.
			//
			// A JSON-RPC 2.0 notification is a request with "json-rpc":"2.0", and
			// without an "id" member. The specification states that notifications
			// must not be responded to. JSON-RPC 2.0 permits the null value as a
			// valid request id, therefore such requests are not notifications.
			//
			// Bitcoin Core serves requests with "id":null or even an absent "id",
			// and responds to such requests with "id":null in the response.
			//
			// Btcd does not respond to any request without and "id" or "id":null,
			// regardless the indicated JSON-RPC protocol version unless RPC quirks
			// are enabled. With RPC quirks enabled, such requests will be responded
			// to if the reqeust does not indicate JSON-RPC version.
			//
			// RPC quirks can be enabled by the user to avoid compatibility issues
			// with software relying on Core's behavior.
			if req.ID == nil && !(cfg.RPCQuirks && req.Jsonrpc == "") {
				return
			}
			resp = s.processRequest(&req, isAdmin, closeChan)
		}

		if resp != nil {
			results = append(results, resp)
		}
	}

	// Process a batched request
	if batchedRequest {
		var batchedRequests []interface{}
		var resp json.RawMessage
		err = json.Unmarshal(body, &batchedRequests)
		if err != nil {
			jsonErr := &RPCError{
				Code: ErrRPCParse.Code,
				Message: fmt.Sprintf("Failed to parse request: %v",
					err),
			}
			resp, err = MarshalResponse(RpcVersion2, nil, nil, jsonErr)
			if err != nil {
				rpcsLog.Errorf("Failed to create reply: %v", err)
			}

			if resp != nil {
				results = append(results, resp)
			}
		}

		if err == nil {
			// Response with an empty batch error if the batch size is zero
			if len(batchedRequests) == 0 {
				jsonErr := &RPCError{
					Code:    ErrRPCInvalidRequest.Code,
					Message: "Invalid request: empty batch",
				}
				resp, err = MarshalResponse(RpcVersion2, nil, nil, jsonErr)
				if err != nil {
					rpcsLog.Errorf("Failed to marshal reply: %v", err)
				}

				if resp != nil {
					results = append(results, resp)
				}
			}

			// Process each batch entry individually
			if len(batchedRequests) > 0 {
				batchSize = len(batchedRequests)

				for _, entry := range batchedRequests {
					var reqBytes []byte
					reqBytes, err = json.Marshal(entry)
					if err != nil {
						jsonErr := &RPCError{
							Code: ErrRPCInvalidRequest.Code,
							Message: fmt.Sprintf("Invalid request: %v",
								err),
						}
						resp, err = MarshalResponse(RpcVersion2, nil, nil, jsonErr)
						if err != nil {
							rpcsLog.Errorf("Failed to create reply: %v", err)
						}

						if resp != nil {
							results = append(results, resp)
						}
						continue
					}

					var req Request
					err := json.Unmarshal(reqBytes, &req)
					if err != nil {
						jsonErr := &RPCError{
							Code: ErrRPCInvalidRequest.Code,
							Message: fmt.Sprintf("Invalid request: %v",
								err),
						}
						resp, err = MarshalResponse("", nil, nil, jsonErr)
						if err != nil {
							rpcsLog.Errorf("Failed to create reply: %v", err)
						}

						if resp != nil {
							results = append(results, resp)
						}
						continue
					}

					resp = s.processRequest(&req, isAdmin, closeChan)
					if resp != nil {
						results = append(results, resp)
					}
				}
			}
		}
	}

	var msg = []byte{}
	if batchedRequest && batchSize > 0 {
		if len(results) > 0 {
			// Form the batched response json
			var buffer bytes.Buffer
			buffer.WriteByte('[')
			for idx, reply := range results {
				if idx == len(results)-1 {
					buffer.Write(reply)
					buffer.WriteByte(']')
					break
				}
				buffer.Write(reply)
				buffer.WriteByte(',')
			}
			msg = buffer.Bytes()
		}
	}

	if !batchedRequest || batchSize == 0 {
		// Respond with the first results entry for single requests
		if len(results) > 0 {
			msg = results[0]
		}
	}

	// Write the response.
	err = s.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		rpcsLog.Error(err)
		return
	}
	if _, err := buf.Write(msg); err != nil {
		rpcsLog.Errorf("Failed to write marshalled reply: %v", err)
	}

	// Terminate with newline to maintain compatibility with Bitcoin Core.
	if err := buf.WriteByte('\n'); err != nil {
		rpcsLog.Errorf("Failed to append terminating newline to reply: %v", err)
	}
}

// Start is used by server.go to start the rpc listener.
func (s *RpcServer) Start() {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	})

	rpcsLog.Trace("Starting RPC server")
	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Handler: rpcServeMux,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}
	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true

		// Limit the number of connections to max allowed.
		if s.limitConnections(w, r.RemoteAddr) {
			return
		}

		// Keep track of the number of connected clients.
		s.incrementClients()
		defer s.decrementClients()

		// Read and respond to the request.
		s.jsonRPCRead(w, r, true)
	})

	for _, listener := range s.cfg.Listeners {
		s.wg.Add(1)
		go func(listener net.Listener) {
			rpcsLog.Infof("RPC server listening on %s", listener.Addr())
			_ = httpServer.Serve(listener)
			rpcsLog.Tracef("RPC listener done for %s", listener.Addr())
			s.wg.Done()
		}(listener)
	}
}

// RpcServerConfig is a descriptor containing the RPC server configuration.
type RpcServerConfig struct {
	// Listeners defines a slice of listeners for which the RPC server will
	// take ownership of and accept connections.  Since the RPC server takes
	// ownership of these listeners, they will be closed when the RPC server
	// is stopped.
	Listeners []net.Listener

	// StartupTime is the unix timestamp for when the server that is hosting
	// the RPC server started.
	StartupTime int64
}

// NewRPCServer returns a new instance of the RpcServer struct.
func NewRPCServer(dbc *storage.DBClient) (*RpcServer, error) {
	//load cfg
	loadCfg()

	//init xylog
	//l, err := logrus.ParseLevel(cfg.DebugLevel)
	//if err != nil {
	//	log.Fatalf("debug_level err:%v", err)
	//}
	//logrus.SetLevel(l)

	// Setup listeners for the configured RPC listen addresses and
	// TLS settings.
	rpcListeners, err := setupRPCListeners()
	if err != nil {
		return nil, fmt.Errorf("setup rpc listens, err:%v", err)
	}

	if len(rpcListeners) == 0 {
		return nil, errors.New("no valid listen address")
	}

	rpc := RpcServer{
		cfg: RpcServerConfig{
			Listeners:   rpcListeners,
			StartupTime: time.Now().Unix(),
		},
		statusLines:            make(map[int]string),
		requestProcessShutdown: make(chan struct{}),
		quit:                   make(chan int),
		dbc:                    dbc,
	}
	if cfg.RPCUser != "" && cfg.RPCPass != "" {
		login := cfg.RPCUser + ":" + cfg.RPCPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpc.authsha = sha256.Sum256([]byte(auth))
	}
	if cfg.RPCLimitUser != "" && cfg.RPCLimitPass != "" {
		login := cfg.RPCLimitUser + ":" + cfg.RPCLimitPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpc.limitauthsha = sha256.Sum256([]byte(auth))
	}
	return &rpc, nil
}

func loadCfg() {
	// Default config.
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}

	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("Loading config: %s", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatalf("open config file[%s] error[%v]", configFileName, err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(cfg); err != nil {
		log.Fatal("load config error: ", err.Error())
	}
}

// setupRPCListeners returns a slice of listeners that are configured for use
// with the RPC server depending on the configuration settings for listen
// addresses and TLS.
func setupRPCListeners() ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen
	netAddrs, err := parseListeners(cfg.RPCListeners)
	if err != nil {
		return nil, err
	}

	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			rpcsLog.Warnf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}

	return listeners, nil
}

// simpleAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}

// String returns the address.
//
// This is part of the net.Addr interface.
func (a simpleAddr) String() string {
	return a.addr
}

// Network returns the network.
//
// This is part of the net.Addr interface.
func (a simpleAddr) Network() string {
	return a.net
}

// Ensure simpleAddr implements the net.Addr interface.
var _ net.Addr = simpleAddr{}

// parseListeners determines whether each listen address is IPv4 and IPv6 and
// returns a slice of appropriate net.Addrs to listen on with TCP. It also
// properly detects addresses which apply to "all interfaces" and adds the
// address as both IPv4 and IPv6.
func parseListeners(addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
			continue
		}

		// Strip IPv6 zone id if present since net.ParseIP does not
		// handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}

		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}

		// To4 returns nil when the IP is not an IPv4 address, so use
		// this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}

func init() {
	rpcHandlers = rpcHandlersBeforeInit
}
