// Copyright (c) 2023-2024 The UXUY Developer Team
// License:
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE

package storage

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/uxuycom/indexer/config"
	"github.com/uxuycom/indexer/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"math/big"
	"reflect"
	"strings"
	"time"
)

const (
	DatabaseTypeSqlite3 = "sqlite3"
	DatabaseTypeMysql   = "mysql"
)

const DBSessionLockKey = "db_session_global_lock_tx"

const (
	OrderByModeAsc  = 1
	OrderByModeDesc = 2
)

const (
	SortTypeId         = 0
	SortTypeDeployTime = 1
	SortTpyeProgress   = 2
	SortTypeHolders    = 3
	SortTypeTxCnt      = 4
)

type DBClient struct {
	SqlDB *gorm.DB
}

// NewDbClient creates a new database client instance.
func NewDbClient(cfg *config.DatabaseConfig) (*DBClient, error) {
	gormCfg := &gorm.Config{}
	if cfg.EnableLog {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}
	switch cfg.Type {
	case DatabaseTypeSqlite3:
		return NewSqliteClient(cfg, gormCfg)
	case DatabaseTypeMysql:
		return NewMysqlClient(cfg, gormCfg)
	}
	return nil, nil
}

func (conn *DBClient) CreateInBatches(dbTx *gorm.DB, value interface{}, batchSize int) error {
	reflectValue := reflect.Indirect(reflect.ValueOf(value))

	// the reflection type judgment of the optimized value
	if reflectValue.Kind() != reflect.Slice && reflectValue.Kind() != reflect.Array {
		return errors.New("value should be slice or array")
	}

	// the reflection length judgment of the optimized value
	reflectLen := reflectValue.Len()
	for i := 0; i < reflectLen; i += batchSize {
		ends := i + batchSize
		if ends > reflectLen {
			ends = reflectLen
		}

		subTx := dbTx.Create(reflectValue.Slice(i, ends).Interface())
		if subTx.Error != nil {
			return subTx.Error
		}
	}
	return nil
}

func (conn *DBClient) SaveLastBlock(tx *gorm.DB, status *model.BlockStatus) error {
	if tx == nil {
		return errors.New("gorm db is not valid")
	}
	return tx.Where("chain = ?", status.Chain).Save(status).Error
}

func (conn *DBClient) QueryLastBlock(chain string) (*big.Int, error) {
	var blockNumberStr string
	err := conn.SqlDB.Table(model.BlockStatus{}.TableName()).Where("chain = ?", chain).Pluck("block_number", &blockNumberStr).Error
	if err != nil {
		return nil, err
	}

	if blockNumberStr == "" {
		return big.NewInt(0), nil
	}

	blockNumber, _ := big.NewInt(0).SetString(blockNumberStr, 10)
	return blockNumber, nil
}

func (conn *DBClient) GetLock() (ok bool, err error) {
	locked := int64(0)
	err = conn.SqlDB.Table(model.BlockStatus{}.TableName()).Raw("SELECT GET_LOCK(?, 0)", DBSessionLockKey).Scan(&locked).Error
	if err != nil {
		return false, err
	}
	return locked > 0, nil
}

type CountResult struct {
	Count int64 `gorm:"column:cnt"`
}

func (conn *DBClient) ReleaseLock() (cnt int64, err error) {
	ret := &CountResult{}
	err = conn.SqlDB.Table(model.BlockStatus{}.TableName()).Raw("SELECT RELEASE_LOCK(?) AS cnt", DBSessionLockKey).Take(ret).Error
	if err != nil {
		return 0, err
	}
	return ret.Count, nil
}

func (conn *DBClient) BatchAddInscription(dbTx *gorm.DB, ins []*model.Inscriptions) error {
	if len(ins) < 1 {
		return nil
	}
	return dbTx.Create(ins).Error
}

func (conn *DBClient) BatchUpdateInscription(dbTx *gorm.DB, chain string, items []*model.Inscriptions) error {
	if len(items) < 1 {
		return nil
	}
	fields := map[string]string{
		"transfer_type": "%d",
	}

	vals := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		vals = append(vals, map[string]interface{}{
			"sid":           item.SID,
			"transfer_type": item.TransferType,
		})
	}
	err, _ := conn.BatchUpdatesBySID(dbTx, chain, model.Inscriptions{}.TableName(), fields, vals)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DBClient) BatchUpdatesBySID(dbTx *gorm.DB, chain string, tblName string, fields map[string]string, values []map[string]interface{}) (error, int64) {
	if len(values) < 1 {
		return nil, 0
	}

	updates := make([]string, 0, len(fields))
	for field, vt := range fields {
		update := fmt.Sprintf(" %s = CASE sid ", field)
		tpl := fmt.Sprintf(" WHEN %s THEN '%s'", "%d", vt)
		for _, value := range values {
			update += fmt.Sprintf(tpl, value["sid"], value[field])
		}
		update += " END"
		updates = append(updates, update)
	}

	ids := make([]string, 0, len(values))
	for _, value := range values {
		ids = append(ids, fmt.Sprintf("%d", value["sid"]))
	}

	finalSql := fmt.Sprintf("UPDATE %s SET %s WHERE chain = '%s' AND sid IN (%s)", tblName, strings.Join(updates, ","), chain, strings.Join(ids, ","))
	ret := dbTx.Exec(finalSql)
	if ret.Error != nil {
		return ret.Error, 0
	}
	return nil, ret.RowsAffected
}

func (conn *DBClient) BatchUpdateInscriptionStats(dbTx *gorm.DB, chain string, items []*model.InscriptionsStats) error {
	if len(items) < 1 {
		return nil
	}

	fields := map[string]string{
		"minted":  "%s",
		"holders": "%d",
		"tx_cnt":  "%d",
	}

	vals := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		vals = append(vals, map[string]interface{}{
			"sid":     item.SID,
			"minted":  item.Minted,
			"holders": item.Holders,
			"tx_cnt":  item.TxCnt,
		})
	}
	err, _ := conn.BatchUpdatesBySID(dbTx, chain, model.InscriptionsStats{}.TableName(), fields, vals)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DBClient) BatchAddInscriptionStats(dbTx *gorm.DB, ins []*model.InscriptionsStats) error {
	if len(ins) < 1 {
		return nil
	}
	return dbTx.Create(ins).Error
}

func (conn *DBClient) BatchAddTransaction(dbTx *gorm.DB, items []*model.Transaction) error {
	if len(items) < 1 {
		return nil
	}
	return conn.CreateInBatches(dbTx, items, 5000)
}

func (conn *DBClient) BatchAddBalanceTx(dbTx *gorm.DB, items []*model.BalanceTxn) error {
	if len(items) < 1 {
		return nil
	}
	return conn.CreateInBatches(dbTx, items, 5000)
}

func (conn *DBClient) BatchAddAddressTx(dbTx *gorm.DB, items []*model.AddressTxs) error {
	if len(items) < 1 {
		return nil
	}
	return conn.CreateInBatches(dbTx, items, 5000)
}

func (conn *DBClient) BatchAddBalances(dbTx *gorm.DB, items []*model.Balances) error {
	if len(items) < 1 {
		return nil
	}
	return conn.CreateInBatches(dbTx, items, 2000)
}

func (conn *DBClient) BatchUpdateBalances(dbTx *gorm.DB, chain string, items []*model.Balances) error {
	if len(items) < 1 {
		return nil
	}

	fields := map[string]string{
		"available": "%s",
		"balance":   "%s",
	}

	vals := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		vals = append(vals, map[string]interface{}{
			"sid":       item.SID,
			"available": item.Available,
			"balance":   item.Balance,
		})
	}
	err, _ := conn.BatchUpdatesBySID(dbTx, chain, model.Balances{}.TableName(), fields, vals)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DBClient) UpdateInscriptionsStatsBySID(dbTx *gorm.DB, chain string, id uint32, updates map[string]interface{}) error {
	return dbTx.Table(model.InscriptionsStats{}.TableName()).Where("chain = ?", chain).Where("sid = ?", id).Updates(updates).Error
}

// FindInscriptionByTick find token by tick
func (conn *DBClient) FindInscriptionByTick(chain, protocol, tick string) (*model.Inscriptions, error) {
	inscriptionBaseInfo := &model.Inscriptions{}
	err := conn.SqlDB.First(inscriptionBaseInfo, "chain = ? AND protocol = ? AND tick = ?", chain, protocol, tick).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return inscriptionBaseInfo, nil
}

// FindInscriptionStatsInfoByBaseId find inscription stats info by base id
func (conn *DBClient) FindInscriptionStatsInfoByBaseId(insId uint32) (*model.InscriptionsStats, error) {
	inscriptionStats := &model.InscriptionsStats{}
	err := conn.SqlDB.First(inscriptionStats, "ins_id = ?", insId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return inscriptionStats, nil
}

func (conn *DBClient) FindUserBalanceByTick(chain, protocol, tick, addr string) (*model.Balances, error) {
	balance := &model.Balances{}
	err := conn.SqlDB.First(balance, "chain = ? AND protocol = ? AND tick = ? AND address = ?", chain, protocol, tick, addr).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return balance, nil
}

func (conn *DBClient) FindTransaction(chain string, hash common.Hash) (*model.Transaction, error) {
	txn := &model.Transaction{}
	err := conn.SqlDB.First(txn, "chain = ? AND tx_hash = ?", chain, hash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return txn, nil
}

func (conn *DBClient) GetInscriptions(limit, offset int, chain, protocol, tick, deployBy string, sort int, sortMode int) (
	[]*model.InscriptionOverView, int64, error) {

	var data []*model.InscriptionOverView
	var total int64

	query := conn.SqlDB.Select("*, (d.minted / a.total_supply) as progress").Table("inscriptions as a").
		Joins("left join `inscriptions_stats` as d on (`a`.chain = `d`.chain and `a`.protocol = `d`.protocol and `a`.tick = `d`.tick)")
	if chain != "" {
		query = query.Where("`a`.chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`a`.protocol = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`a`.tick = ?", tick)
	}
	if deployBy != "" {
		query = query.Where("`a`.deploy_by = ?", deployBy)
	}

	// sort mode 1: asc 2: desc
	mode := "desc"
	if sortMode == OrderByModeAsc {
		mode = "asc"
	}

	// sort by  0.id  1.deploy_time  2.progress  3.holders  4.tx_cnt
	switch sort {
	case SortTypeId:
		query = query.Order("`a`.id " + mode)
	case SortTypeDeployTime:
		query = query.Order("deploy_time " + mode)
	case SortTpyeProgress:
		query = query.Order("progress " + mode)
	case SortTypeHolders:
		query = query.Order("holders " + mode)
	case SortTypeTxCnt:
		query = query.Order("tx_cnt " + mode)
	}

	query = query.Count(&total)
	result := query.Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) FindInscriptionInfo(chain, protocol, tick, deployHash string) (*model.InscriptionOverView, error) {
	var inscription model.InscriptionOverView
	result := conn.SqlDB.Model(&model.Inscriptions{}).
		Select("inscriptions.*, inscriptions_stats.*, (inscriptions_stats.minted / inscriptions.total_supply) as progress").
		Joins("left join inscriptions_stats ON inscriptions.chain = inscriptions_stats.chain AND inscriptions.protocol = inscriptions_stats.protocol AND inscriptions.tick = inscriptions_stats.tick")

	if chain != "" {
		result = result.Where("inscriptions.chain = ?", chain)
	}
	if protocol != "" {
		result = result.Where("inscriptions.protocol = ?", protocol)
	}
	if tick != "" {
		result = result.Where("inscriptions.tick = ?", tick)
	}
	if deployHash != "" {
		result = result.Where("inscriptions.deploy_hash = ?", deployHash)
	}

	err := result.First(&inscription).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &inscription, nil
}

func (conn *DBClient) GetInscriptionsByIdLimit(chain string, start uint64, limit int) ([]model.Inscriptions, error) {
	inscriptions := make([]model.Inscriptions, 0)
	err := conn.SqlDB.Where("chain = ?", chain).Where("id > ?", start).Order("id asc").Limit(limit).Find(&inscriptions).Error
	if err != nil {
		return nil, err
	}
	return inscriptions, nil
}

func (conn *DBClient) GetInscriptionStatsByIdLimit(chain string, start uint64, limit int) ([]model.InscriptionsStats, error) {
	stats := make([]model.InscriptionsStats, 0)
	err := conn.SqlDB.Where("chain = ?", chain).Where("id > ?", start).Order("id asc").Limit(limit).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (conn *DBClient) GetInscriptionStats(chain string, start uint64, limit int) ([]model.InscriptionsStats, error) {
	stats := make([]model.InscriptionsStats, 0)
	err := conn.SqlDB.Where("chain = ?", chain).Where("id > ?", start).Order("id asc").Limit(limit).Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (conn *DBClient) GetInscriptionStatsList(limit int, offset int, sort int) ([]model.InscriptionsStats, int64, error) {
	stats := make([]model.InscriptionsStats, 0)
	query := conn.SqlDB.Model(&model.InscriptionsStats{})

	var total int64
	query.Count(&total)

	orderBy := " id DESC"
	if sort == OrderByModeAsc {
		orderBy = " id ASC"
	}
	err := query.Order(orderBy).Limit(limit).Offset(offset).Find(&stats).Error
	if err != nil {
		return nil, 0, err
	}
	return stats, total, nil
}

func (conn *DBClient) GetInscriptionsByAddress(limit, offset int, address string) ([]*model.Balances, error) {
	balances := make([]*model.Balances, 0)

	query := conn.SqlDB.Model(&model.Balances{})
	if address != "" {
		query = query.Where("`address` = ?", address)
	}

	result := query.Order("id desc").Limit(limit).Offset(offset).Find(&balances)
	if result.Error != nil {
		return nil, result.Error
	}

	return balances, nil
}

func (conn *DBClient) GetTransactionsByAddress(limit, offset int, address, chain, protocol, tick, key string, event int8) (
	[]*model.AddressTransaction, int64, error) {

	var data []*model.AddressTransaction
	var total int64

	query := conn.SqlDB.Select("*").Table("txs as t").
		Joins("left join `address_txs` as a on (`t`.tx_hash = `a`.tx_hash and `t`.chain = `a`.chain and `t`.protocol = `a`.protocol and `t`.tick = `a`.tick)").
		Where("`a`.address = ?", address)

	if chain != "" {
		query = query.Where("`a`.chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`a`.protocol = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`a`.tick = ?", tick)
	}
	if key != "" {
		query = query.Where("`a`.tick like ?", "%"+key+"%")
	}
	if event > 0 {
		query = query.Where("`a`.event = ?", event)
	}

	query = query.Count(&total)
	result := query.Order("`a`.id desc").Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) GetAddressTxs(limit, offset int, address, chain, protocol, tick string, event int8) ([]*model.AddressTransaction, int64, error) {
	var data []*model.AddressTransaction
	var total int64

	query := conn.SqlDB.Select("*").Table("`address_txs`").
		Where("address = ?", address)

	if chain != "" {
		query = query.Where("chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("protocol = ?", protocol)
	}
	if tick != "" {
		//query = query.Where("tick = ?", tick)
		query.Where("tick like ?", "%"+tick+"%")
	}
	if event > 0 {
		query = query.Where("event = ?", event)
	}

	query = query.Count(&total)
	result := query.Order("id desc").Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) GetTxsByHashes(chain string, hashes []common.Hash) ([]*model.Transaction, error) {
	txs := make([]*model.Transaction, 0)
	err := conn.SqlDB.Where("chain = ? AND tx_hash in ?", chain, hashes).Find(&txs).Error
	if err != nil {
		return nil, err
	}
	return txs, nil
}

// GetTransactions find all transaction
func (conn *DBClient) GetTransactions(chain string, address string, tick string, limit int, offset int, sort int) ([]*model.Transaction, int64, error) {

	txs := make([]*model.Transaction, 0)
	query := conn.SqlDB.Model(&model.Transaction{})

	var total int64

	if len(chain) > 0 {
		query = query.Where("chain = ?", chain)
	}

	if len(tick) > 0 {
		query = query.Where("tick = ?", tick)
	}

	if len(address) > 0 {
		query = query.Where("from = ? or to = ?", address, address)
	}

	orderBy := " id DESC"
	if sort == OrderByModeAsc {
		orderBy = " id ASC"
	}
	err := query.Order(orderBy).Limit(limit).Offset(offset).Find(&txs).Error
	if err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

func (conn *DBClient) GetAddressInscriptions(limit, offset int, address, chain, protocol, tick string,
	key string, sort int) (
	[]*model.BalanceInscription, int64, error) {

	var data []*model.BalanceInscription
	var total int64

	query := conn.SqlDB.Select("*").Table("balances as b").
		Joins("left join `inscriptions` as a on (`b`.chain = `a`.chain and `b`.protocol = `a`.protocol and `b`.tick = `a`.tick)")

	query = query.Where("`b`.address = ? and `b`.balance > 0", address)

	if chain != "" {
		query = query.Where("`b`.chain = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`b`.protocol = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`b`.tick = ?", tick)
	}
	if key != "" {
		query = query.Where("`b`.tick like ?", "%"+key+"%")
	}

	query = query.Count(&total)
	orderBy := "`b`.balance DESC"
	if sort == OrderByModeAsc {
		orderBy = "`b`.balance ASC"
	}

	result := query.Order(orderBy).Limit(limit).Offset(offset).Find(&data)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return data, total, nil
}

func (conn *DBClient) GetBalancesChainByAddress(limit, offset int, address, chain, protocol, tick string) (
	[]*model.BalanceChain, int64, error) {

	var balances []*model.BalanceChain
	var total int64

	query := conn.SqlDB.Select("chain,address,SUM(balance) as balance").Table("balances").Where("`address` = ?", address)
	if chain != "" {
		query = query.Where("`chain` = ?", chain)
	}
	if protocol != "" {
		query = query.Where("`protocol` = ?", protocol)
	}
	if tick != "" {
		query = query.Where("`tick` = ?", tick)
	}
	query = query.Count(&total)
	orderBy := "balance DESC"
	groupBy := "chain"
	err := query.Group(groupBy).Order(orderBy).Limit(limit).Offset(offset).Find(&balances).Error
	if err != nil {
		return nil, 0, err
	}
	return balances, total, nil
}

func (conn *DBClient) GetHoldersByTick(limit, offset int, chain, protocol, tick string, sortMode int) ([]*model.Balances, int64, error) {
	var holders []*model.Balances
	var total int64
	query := conn.SqlDB.Model(&model.Balances{}).
		Where("balance > 0 and chain = ? and protocol = ? and tick = ?", chain, protocol, tick)
	query = query.Count(&total)
	orderBy := "balance desc,"
	if sortMode == OrderByModeAsc {
		orderBy = "balance asc,"
	}

	result := query.Order(orderBy + " id asc").Limit(limit).Offset(offset).Find(&holders)
	if result.Error != nil {
		return nil, 0, result.Error
	}
	return holders, total, nil
}

func (conn *DBClient) GetUTXOCount(address, chain, protocol, tick string) (int64, error) {
	var count int64
	query := conn.SqlDB.Model(&model.UTXO{}).
		Where("address = ? and chain = ? and protocol = ? and tick = ? and status = ?", address, chain, protocol, tick, model.UTXOStatusUnspent)
	err := query.Count(&count)
	if err.Error != nil {
		return 0, err.Error
	}
	return count, nil
}

func (conn *DBClient) GetBalancesByIdLimit(chain string, start uint64, limit int) ([]model.Balances, error) {
	balances := make([]model.Balances, 0)
	err := conn.SqlDB.Where("chain = ?", chain).Where("id > ?", start).Order("id asc").Limit(limit).Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return balances, nil
}

func (conn *DBClient) GetUTXOsByIdLimit(start uint64, limit int) ([]model.UTXO, error) {
	utxos := make([]model.UTXO, 0, limit)
	err := conn.SqlDB.Where("id > ? ", start).Where("status = ? ", model.UTXOStatusUnspent).Order("id asc").Limit(limit).Find(&utxos).Error
	if err != nil {
		return nil, err
	}
	return utxos, nil
}

func (conn *DBClient) GetUtxosByAddress(address, chain, protocol, tick string) ([]*model.UTXO, error) {
	var utxos []*model.UTXO
	query := conn.SqlDB.Model(&model.UTXO{}).
		Where("address = ? and chain = ? and protocol = ? and tick = ? and status = ?", address, chain, protocol, tick, model.UTXOStatusUnspent)
	result := query.Order("id desc").Find(&utxos)
	if result.Error != nil {
		return nil, result.Error
	}
	return utxos, nil
}

func (conn *DBClient) FindAddressTxByHash(chain string, hash common.Hash) (*model.AddressTxs, error) {
	tx := &model.AddressTxs{}
	err := conn.SqlDB.First(tx, "chain = ? and tx_hash = ? ", chain, hash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return tx, nil
}

func (conn *DBClient) FindBalanceByTxHash(hash string) ([]*model.BalanceTxn, error) {
	balances := make([]*model.BalanceTxn, 0)
	str := "SELECT * FROM balance_txn WHERE tx_hash = " + hash
	err := conn.SqlDB.Raw(str).Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return balances, nil
}

// GetAllChainFromBlock query all chains from block table
func (conn *DBClient) GetAllChainFromBlock() ([]string, error) {
	var chains []string
	err := conn.SqlDB.Model(&model.Block{}).Distinct().Pluck("chain", &chains).Error
	if err != nil {
		return nil, err
	}
	return chains, nil
}

// GetAllBlocks query all last block from block table
func (conn *DBClient) GetAllBlocks() ([]model.Block, error) {
	blocks := make([]model.Block, 0)
	err := conn.SqlDB.Model(&model.Block{}).Find(&blocks).Error
	if err != nil {
		return nil, err
	}
	return blocks, nil
}

func (conn *DBClient) FindLastBlock(chain string) (*model.Block, error) {
	data := &model.Block{}
	err := conn.SqlDB.First(data, "chain = ? ", chain).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (conn *DBClient) GetInscriptionsByChain(chain string, hashes []string) ([]*model.Inscriptions, error) {
	inscriptions := make([]*model.Inscriptions, 0)
	err := conn.SqlDB.Where("chain = ? AND deploy_hash in ?", chain, hashes).Find(&inscriptions).Error
	if err != nil {
		return nil, err
	}
	return inscriptions, nil
}

func (conn *DBClient) FindInscriptionsStatsByTick(chain string, protocol string, tick string) (*model.InscriptionsStats, error) {
	inscriptionStats := &model.InscriptionsStats{}
	err := conn.SqlDB.First(inscriptionStats, "chain = ? AND protocol = ? AND tick = ?", chain, protocol, tick).Error
	if err != nil {
		return nil, err
	}

	return inscriptionStats, nil
}

func (conn *DBClient) FindLastChainStatHourByChainAndDateHour(chain string, dateHour uint32) (*model.ChainStatHour, error) {
	data := &model.ChainStatHour{}
	err := conn.SqlDB.Last(data, "chain = ?", chain).Where("date_hour = ?", dateHour).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (conn *DBClient) FindAddressTxByIdAndChainAndLimit(chain string, start uint64, limit int) ([]model.AddressTxs, error) {
	txs := make([]model.AddressTxs, 0)
	err := conn.SqlDB.Where("id > ?", start).Where("chain = ?", chain).Order("id asc").Limit(limit).Find(&txs).Error
	if err != nil {
		return nil, err
	}
	return txs, nil
}

func (conn *DBClient) FindInscriptionsTxByIdAndChainAndLimit(chain string, nowHour, lastHour time.Time) ([]model.Inscriptions, error) {
	inscriptions := make([]model.Inscriptions, 0)
	err := conn.SqlDB.Where("chain = ?", chain).Where("deploy_time > ? and deploy_time < ?", lastHour, nowHour).Find(&inscriptions).Error
	if err != nil {
		return nil, err
	}
	return inscriptions, nil
}

func (conn *DBClient) FindBalanceTxByIdAndChainAndLimit(chain string, balanceIndex uint64, limit int) ([]model.BalanceTxn, error) {
	balances := make([]model.BalanceTxn, 0)
	err := conn.SqlDB.Where("id > ?", balanceIndex).Where("chain = ?", chain).Where("amount > 0").Order("id asc").Limit(limit).Find(&balances).Error
	if err != nil {
		return nil, err
	}
	return balances, nil
}
func (conn *DBClient) AddChainStatHour(chainStatHour *model.ChainStatHour) error {
	return conn.SqlDB.Create(chainStatHour).Error
}

func (conn *DBClient) GetAllChainInfo() ([]model.ChainInfo, error) {
	chains := make([]model.ChainInfo, 0)
	err := conn.SqlDB.Model(&model.ChainInfo{}).Find(&chains).Error
	if err != nil {
		return nil, err
	}
	return chains, nil
}

func (conn *DBClient) GetChainInfoByChain(chain string) (*model.ChainInfo, error) {
	chainInfo := &model.ChainInfo{}
	err := conn.SqlDB.Model(&model.ChainInfo{}).Where("chain = ?", chain).Find(&chainInfo).Error
	if err != nil {
		return nil, err
	}
	return chainInfo, nil
}

func (conn *DBClient) GroupChainStatHourBy24Hour(startHour, endHour uint32, chain []string) ([]model.GroupChainStatHour, error) {
	stats := make([]model.GroupChainStatHour, 0)
	tx := conn.SqlDB.Select("chain,SUM(address_count) as address_count,SUM(inscriptions_count) as inscriptions_count,SUM(balance_sum) as balance_sum").
		Where("date_hour >= ? and date_hour <= ?", endHour, startHour)
	if len(chain) > 0 {
		tx = tx.Where("chain in ?", chain)
	}
	err := tx.Table("chain_stats_hour").Group("chain").Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (conn *DBClient) GroupChainBlockStat(startId uint64, chain string) ([]model.ChainBlockStat, error) {
	stats := make([]model.ChainBlockStat, 0)
	tx := conn.SqlDB.Select("block_height,count(distinct(tick)) as tick_count,count(*) as transaction_count,min(created_at) as created_at").
		Where("id> ? and chain = ?", startId, chain)
	err := tx.Table("txs").Group("chain,block_height").Limit(10).Order("created_at desc").Find(&stats).Error
	if err != nil {
		return nil, err
	}
	return stats, nil
}
func (conn *DBClient) MaxIdFromTransaction() (uint64, error) {
	var id uint64
	err := conn.SqlDB.Model(&model.Transaction{}).Select("max(id)").Scan(&id).Error
	if err != nil {
		return 0, err
	}
	return id, nil
}
