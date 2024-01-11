package protocol

import (
	"encoding/hex"
	"open-indexer/devents"
	"open-indexer/model"
	"reflect"
	"testing"
)

func TestParseEVMMetaData(t *testing.T) {
	type args struct {
		chain     string
		inputData string
	}

	inputData := hex.EncodeToString([]byte(",{\"p\":\"asc-20\",\"op\":\"deploy\",\"tick\":\"Tduck\",\"max\":\"210000000\",\"lim\":\"1000\"}"))
	tests := []struct {
		name    string
		args    args
		want    *devents.MetaData
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Can be passed",
			args: args{
				chain:     model.ChainAVAX,
				inputData: "0x646174613a2c7b2270223a226173632d3230222c226f70223a226465706c6f79222c227469636b223a22546475636b222c226d6178223a22323130303030303030222c226c696d223a2231303030227d",
			},
			want: &devents.MetaData{
				Chain:    model.ChainAVAX,
				Operate:  "deploy",
				Protocol: "asc-20",
				Tick:     "tduck",
				Data:     "{\"p\":\"asc-20\",\"op\":\"deploy\",\"tick\":\"Tduck\",\"max\":\"210000000\",\"lim\":\"1000\"}",
			},
			wantErr: false,
		},
		{
			name: "Can be passed",
			args: args{
				chain:     model.ChainAVAX,
				inputData: "0x" + inputData,
			},
			want: &devents.MetaData{
				Chain:    model.ChainAVAX,
				Operate:  "deploy",
				Protocol: "asc-20",
				Tick:     "tduck",
				Data:     "{\"p\":\"asc-20\",\"op\":\"deploy\",\"tick\":\"Tduck\",\"max\":\"210000000\",\"lim\":\"1000\"}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEVMMetaData(tt.args.chain, tt.args.inputData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEVMMetaData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEVMMetaData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
