package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prysmaticlabs/go-bitfield"
	v1 "github.com/prysmaticlabs/prysm/proto/engine/v1"
	eth "github.com/prysmaticlabs/prysm/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/testing/require"
)

func ezDecode(t *testing.T, s string) []byte {
	v, err := hexutil.Decode(s)
	require.NoError(t, err)
	return v
}

func TestSignedValidatorRegistration_MarshalJSON(t *testing.T) {
	svr := &eth.SignedValidatorRegistrationV1{
		Message: &eth.ValidatorRegistrationV1{
			FeeRecipient: make([]byte, 20),
			GasLimit:     0,
			Timestamp:    0,
			Pubkey:       make([]byte, 48),
		},
		Signature: make([]byte, 96),
	}
	je, err := json.Marshal(&SignedValidatorRegistration{SignedValidatorRegistrationV1: svr})
	require.NoError(t, err)
	// decode with a struct w/ plain strings so we can check the string encoding of the hex fields
	un := struct {
		Message struct {
			FeeRecipient string `json:"fee_recipient"`
			Pubkey       string `json:"pubkey"`
		} `json:"message"`
		Signature string `json:"signature"`
	}{}
	require.NoError(t, json.Unmarshal(je, &un))
	require.Equal(t, "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", un.Signature)
	require.Equal(t, "0x0000000000000000000000000000000000000000", un.Message.FeeRecipient)
	require.Equal(t, "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", un.Message.Pubkey)
}

var testExampleHeaderResponse = `{
  "version": "bellatrix",
  "data": {
    "message": {
      "header": {
        "parent_hash": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
        "fee_recipient": "0xabcf8e0d4e9587369b2301d0790347320302cc09",
        "state_root": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
        "receipts_root": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
        "logs_bloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        "prev_randao": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
        "block_number": "1",
        "gas_limit": "1",
        "gas_used": "1",
        "timestamp": "1",
        "extra_data": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
        "base_fee_per_gas": "452312848583266388373324160190187140051835877600158453279131187530910662656",
        "block_hash": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
        "transactions_root": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"
      },
      "value": "652312848583266388373324160190187140051835877600158453279131187530910662656",
      "pubkey": "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"
    },
    "signature": "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"
  }
}`

func TestExecutionHeaderResponseUnmarshal(t *testing.T) {
	hr := &ExecHeaderResponse{}
	require.NoError(t, json.Unmarshal([]byte(testExampleHeaderResponse), hr))
	cases := []struct {
		expected string
		actual   string
		name     string
	}{
		{
			expected: "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505",
			actual:   hexutil.Encode(hr.Data.Signature),
			name:     "Signature",
		},
		{
			expected: "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a",
			actual:   hexutil.Encode(hr.Data.Message.Pubkey),
			name:     "ExecHeaderResponse.Pubkey",
		},
		{
			expected: "652312848583266388373324160190187140051835877600158453279131187530910662656",
			actual:   hr.Data.Message.Value.String(),
			name:     "ExecHeaderResponse.Value",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.ParentHash),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.ParentHash",
		},
		{
			expected: "0xabcf8e0d4e9587369b2301d0790347320302cc09",
			actual:   hexutil.Encode(hr.Data.Message.Header.FeeRecipient),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.FeeRecipient",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.StateRoot),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.StateRoot",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.ReceiptsRoot),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.ReceiptsRoot",
		},
		{
			expected: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			actual:   hexutil.Encode(hr.Data.Message.Header.LogsBloom),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.LogsBloom",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.PrevRandao),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.PrevRandao",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", hr.Data.Message.Header.BlockNumber),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.BlockNumber",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", hr.Data.Message.Header.GasLimit),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.GasLimit",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", hr.Data.Message.Header.GasUsed),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.GasUsed",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", hr.Data.Message.Header.Timestamp),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.Timestamp",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.ExtraData),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.ExtraData",
		},
		{
			expected: "452312848583266388373324160190187140051835877600158453279131187530910662656",
			actual:   fmt.Sprintf("%d", hr.Data.Message.Header.BaseFeePerGas),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.BaseFeePerGas",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.BlockHash),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.BlockHash",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(hr.Data.Message.Header.TransactionsRoot),
			name:     "ExecHeaderResponse.ExecutionPayloadHeader.TransactionsRoot",
		},
	}
	for _, c := range cases {
		require.Equal(t, c.expected, c.actual, fmt.Sprintf("unexpected value for field %s", c.name))
	}
}

func TestExecutionHeaderResponseToProto(t *testing.T) {
	bfpg := stringToUint256("452312848583266388373324160190187140051835877600158453279131187530910662656")
	v := stringToUint256("652312848583266388373324160190187140051835877600158453279131187530910662656")
	hr := &ExecHeaderResponse{}
	require.NoError(t, json.Unmarshal([]byte(testExampleHeaderResponse), hr))
	p, err := hr.ToProto()
	require.NoError(t, err)
	signature, err := hexutil.Decode("0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505")
	require.NoError(t, err)
	pubkey, err := hexutil.Decode("0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a")
	require.NoError(t, err)
	parentHash, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	feeRecipient, err := hexutil.Decode("0xabcf8e0d4e9587369b2301d0790347320302cc09")
	require.NoError(t, err)
	stateRoot, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	receiptsRoot, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	logsBloom, err := hexutil.Decode("0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	prevRandao, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	extraData, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	blockHash, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	txRoot, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)

	expected := &eth.SignedBuilderBid{
		Message: &eth.BuilderBid{
			Header: &eth.ExecutionPayloadHeader{
				ParentHash:       parentHash,
				FeeRecipient:     feeRecipient,
				StateRoot:        stateRoot,
				ReceiptsRoot:     receiptsRoot,
				LogsBloom:        logsBloom,
				PrevRandao:       prevRandao,
				BlockNumber:      1,
				GasLimit:         1,
				GasUsed:          1,
				Timestamp:        1,
				ExtraData:        extraData,
				BaseFeePerGas:    bfpg.SSZBytes(),
				BlockHash:        blockHash,
				TransactionsRoot: txRoot,
			},
			Value:  v.SSZBytes(),
			Pubkey: pubkey,
		},
		Signature: signature,
	}
	require.DeepEqual(t, expected, p)
}

var testExampleExecutionPayload = `{
  "version": "bellatrix",
  "data": {
    "parent_hash": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
    "fee_recipient": "0xabcf8e0d4e9587369b2301d0790347320302cc09",
    "state_root": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
    "receipts_root": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
    "logs_bloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "prev_randao": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
    "block_number": "1",
    "gas_limit": "1",
    "gas_used": "1",
    "timestamp": "1",
    "extra_data": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
    "base_fee_per_gas": "452312848583266388373324160190187140051835877600158453279131187530910662656",
    "block_hash": "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
    "transactions": [
      "0x02f878831469668303f51d843b9ac9f9843b9aca0082520894c93269b73096998db66be0441e836d873535cb9c8894a19041886f000080c001a031cc29234036afbf9a1fb9476b463367cb1f957ac0b919b69bbc798436e604aaa018c4e9c3914eb27aadd0b91e10b18655739fcf8c1fc398763a9f1beecb8ddc86"
    ]
  }
}`

func TestExecutionPayloadResponseUnmarshal(t *testing.T) {
	epr := &ExecPayloadResponse{}
	require.NoError(t, json.Unmarshal([]byte(testExampleExecutionPayload), epr))
	cases := []struct {
		expected string
		actual   string
		name     string
	}{
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(epr.Data.ParentHash),
			name:     "ExecPayloadResponse.ExecutionPayload.ParentHash",
		},
		{
			expected: "0xabcf8e0d4e9587369b2301d0790347320302cc09",
			actual:   hexutil.Encode(epr.Data.FeeRecipient),
			name:     "ExecPayloadResponse.ExecutionPayload.FeeRecipient",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(epr.Data.StateRoot),
			name:     "ExecPayloadResponse.ExecutionPayload.StateRoot",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(epr.Data.ReceiptsRoot),
			name:     "ExecPayloadResponse.ExecutionPayload.ReceiptsRoot",
		},
		{
			expected: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
			actual:   hexutil.Encode(epr.Data.LogsBloom),
			name:     "ExecPayloadResponse.ExecutionPayload.LogsBloom",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(epr.Data.PrevRandao),
			name:     "ExecPayloadResponse.ExecutionPayload.PrevRandao",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", epr.Data.BlockNumber),
			name:     "ExecPayloadResponse.ExecutionPayload.BlockNumber",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", epr.Data.GasLimit),
			name:     "ExecPayloadResponse.ExecutionPayload.GasLimit",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", epr.Data.GasUsed),
			name:     "ExecPayloadResponse.ExecutionPayload.GasUsed",
		},
		{
			expected: "1",
			actual:   fmt.Sprintf("%d", epr.Data.Timestamp),
			name:     "ExecPayloadResponse.ExecutionPayload.Timestamp",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(epr.Data.ExtraData),
			name:     "ExecPayloadResponse.ExecutionPayload.ExtraData",
		},
		{
			expected: "452312848583266388373324160190187140051835877600158453279131187530910662656",
			actual:   fmt.Sprintf("%d", epr.Data.BaseFeePerGas),
			name:     "ExecPayloadResponse.ExecutionPayload.BaseFeePerGas",
		},
		{
			expected: "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2",
			actual:   hexutil.Encode(epr.Data.BlockHash),
			name:     "ExecPayloadResponse.ExecutionPayload.BlockHash",
		},
	}
	for _, c := range cases {
		require.Equal(t, c.expected, c.actual, fmt.Sprintf("unexpected value for field %s", c.name))
	}
	require.Equal(t, 1, len(epr.Data.Transactions))
	txHash := "0x02f878831469668303f51d843b9ac9f9843b9aca0082520894c93269b73096998db66be0441e836d873535cb9c8894a19041886f000080c001a031cc29234036afbf9a1fb9476b463367cb1f957ac0b919b69bbc798436e604aaa018c4e9c3914eb27aadd0b91e10b18655739fcf8c1fc398763a9f1beecb8ddc86"
	require.Equal(t, txHash, hexutil.Encode(epr.Data.Transactions[0]))
}

func TestExecutionPayloadResponseToProto(t *testing.T) {
	hr := &ExecPayloadResponse{}
	require.NoError(t, json.Unmarshal([]byte(testExampleExecutionPayload), hr))
	p, err := hr.ToProto()
	require.NoError(t, err)

	parentHash, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	feeRecipient, err := hexutil.Decode("0xabcf8e0d4e9587369b2301d0790347320302cc09")
	require.NoError(t, err)
	stateRoot, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	receiptsRoot, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	logsBloom, err := hexutil.Decode("0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	prevRandao, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	extraData, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)
	blockHash, err := hexutil.Decode("0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")
	require.NoError(t, err)

	tx, err := hexutil.Decode("0x02f878831469668303f51d843b9ac9f9843b9aca0082520894c93269b73096998db66be0441e836d873535cb9c8894a19041886f000080c001a031cc29234036afbf9a1fb9476b463367cb1f957ac0b919b69bbc798436e604aaa018c4e9c3914eb27aadd0b91e10b18655739fcf8c1fc398763a9f1beecb8ddc86")
	require.NoError(t, err)
	txList := [][]byte{tx}

	bfpg := stringToUint256("452312848583266388373324160190187140051835877600158453279131187530910662656")
	expected := &v1.ExecutionPayload{
		ParentHash:    parentHash,
		FeeRecipient:  feeRecipient,
		StateRoot:     stateRoot,
		ReceiptsRoot:  receiptsRoot,
		LogsBloom:     logsBloom,
		PrevRandao:    prevRandao,
		BlockNumber:   1,
		GasLimit:      1,
		GasUsed:       1,
		Timestamp:     1,
		ExtraData:     extraData,
		BaseFeePerGas: bfpg.SSZBytes(),
		BlockHash:     blockHash,
		Transactions:  txList,
	}
	require.DeepEqual(t, expected, p)
}

func pbEth1Data() *eth.Eth1Data {
	return &eth.Eth1Data{
		DepositRoot:  make([]byte, 32),
		DepositCount: 23,
		BlockHash:    make([]byte, 32),
	}
}

func TestEth1DataMarshal(t *testing.T) {
	ed := &Eth1Data{
		Eth1Data: pbEth1Data(),
	}
	b, err := json.Marshal(ed)
	require.NoError(t, err)
	expected := `{"deposit_root":"0x0000000000000000000000000000000000000000000000000000000000000000","deposit_count":"23","block_hash":"0x0000000000000000000000000000000000000000000000000000000000000000"}`
	require.Equal(t, expected, string(b))
}

func pbSyncAggregate() *eth.SyncAggregate {
	return &eth.SyncAggregate{
		SyncCommitteeSignature: make([]byte, 48),
		SyncCommitteeBits:      bitfield.Bitvector512{0x01},
	}
}

func TestSyncAggregate_MarshalJSON(t *testing.T) {
	sa := &SyncAggregate{pbSyncAggregate()}
	b, err := json.Marshal(sa)
	require.NoError(t, err)
	expected := `{"sync_committee_bits":"0x01","sync_committee_signature":"0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"}`
	require.Equal(t, expected, string(b))
}

func pbDeposit(t *testing.T) *eth.Deposit {
	return &eth.Deposit{
		Proof: [][]byte{ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2")},
		Data: &eth.Deposit_Data{
			PublicKey:             ezDecode(t, "0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a"),
			WithdrawalCredentials: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			Amount:                1,
			Signature:             ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
		},
	}
}

func TestDeposit_MarshalJSON(t *testing.T) {
	d := &Deposit{
		Deposit: pbDeposit(t),
	}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	expected := `{"proof":["0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"],"data":{"pubkey":"0x93247f2209abcacf57b75a51dafae777f9dd38bc7053d1af526f220a7489a6d3a2753e5f3e8b1cfe39b56f43611df74a","withdrawal_credentials":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","amount":"1","signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"}}`
	require.Equal(t, expected, string(b))
}

func pbSignedVoluntaryExit(t *testing.T) *eth.SignedVoluntaryExit {
	return &eth.SignedVoluntaryExit{
		Exit: &eth.VoluntaryExit{
			Epoch:          1,
			ValidatorIndex: 1,
		},
		Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
	}
}

func TestVoluntaryExit(t *testing.T) {
	ve := &SignedVoluntaryExit{
		SignedVoluntaryExit: pbSignedVoluntaryExit(t),
	}
	b, err := json.Marshal(ve)
	require.NoError(t, err)
	expected := `{"message":{"epoch":"1","validator_index":"1"},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"}`
	require.Equal(t, expected, string(b))
}

func pbAttestation(t *testing.T) *eth.Attestation {
	return &eth.Attestation{
		AggregationBits: bitfield.Bitlist{0x01},
		Data: &eth.AttestationData{
			Slot:            1,
			CommitteeIndex:  1,
			BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			Source: &eth.Checkpoint{
				Epoch: 1,
				Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			},
			Target: &eth.Checkpoint{
				Epoch: 1,
				Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			},
		},
		Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
	}
}

func TestAttestationMarshal(t *testing.T) {
	a := &Attestation{
		Attestation: pbAttestation(t),
	}
	b, err := json.Marshal(a)
	require.NoError(t, err)
	expected := `{"aggregation_bits":"0x01","data":{"slot":"1","index":"1","beacon_block_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","source":{"epoch":"1","root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"},"target":{"epoch":"1","root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"}},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"}`
	require.Equal(t, expected, string(b))
}

func pbAttesterSlashing(t *testing.T) *eth.AttesterSlashing {
	return &eth.AttesterSlashing{
		Attestation_1: &eth.IndexedAttestation{
			AttestingIndices: []uint64{1},
			Signature:        ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
			Data: &eth.AttestationData{
				Slot:            1,
				CommitteeIndex:  1,
				BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				Source: &eth.Checkpoint{
					Epoch: 1,
					Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
				Target: &eth.Checkpoint{
					Epoch: 1,
					Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
			},
		},
		Attestation_2: &eth.IndexedAttestation{
			AttestingIndices: []uint64{1},
			Signature:        ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
			Data: &eth.AttestationData{
				Slot:            1,
				CommitteeIndex:  1,
				BeaconBlockRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				Source: &eth.Checkpoint{
					Epoch: 1,
					Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
				Target: &eth.Checkpoint{
					Epoch: 1,
					Root:  ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				},
			},
		},
	}
}

func TestAttesterSlashing_MarshalJSON(t *testing.T) {
	as := &AttesterSlashing{
		AttesterSlashing: pbAttesterSlashing(t),
	}
	b, err := json.Marshal(as)
	require.NoError(t, err)
	expected := `{"attestation_1":{"attesting_indices":["1"],"data":{"slot":"1","index":"1","beacon_block_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","source":{"epoch":"1","root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"},"target":{"epoch":"1","root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"}},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"},"attestation_2":{"attesting_indices":["1"],"data":{"slot":"1","index":"1","beacon_block_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","source":{"epoch":"1","root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"},"target":{"epoch":"1","root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"}},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"}}`
	require.Equal(t, expected, string(b))
}

func pbProposerSlashing(t *testing.T) *eth.ProposerSlashing {
	return &eth.ProposerSlashing{
		Header_1: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 1,
				ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			},
			Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
		},
		Header_2: &eth.SignedBeaconBlockHeader{
			Header: &eth.BeaconBlockHeader{
				Slot:          1,
				ProposerIndex: 1,
				ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
				BodyRoot:      ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
			},
			Signature: ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
		},
	}
}

func TestProposerSlashings(t *testing.T) {
	ps := &ProposerSlashing{ProposerSlashing: pbProposerSlashing(t)}
	b, err := json.Marshal(ps)
	require.NoError(t, err)
	expected := `{"signed_header_1":{"message":{"slot":"1","proposer_index":"1","parent_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","state_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","body_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"},"signed_header_2":{"message":{"slot":"1","proposer_index":"1","parent_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","state_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","body_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"},"signature":"0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"}}`
	require.Equal(t, expected, string(b))
}

func pbExecutionPayloadHeader(t *testing.T) *eth.ExecutionPayloadHeader {
	bfpg := stringToUint256("452312848583266388373324160190187140051835877600158453279131187530910662656")
	return &eth.ExecutionPayloadHeader{
		ParentHash:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		FeeRecipient:     ezDecode(t, "0xabcf8e0d4e9587369b2301d0790347320302cc09"),
		StateRoot:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		ReceiptsRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		LogsBloom:        ezDecode(t, "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		PrevRandao:       ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		BlockNumber:      1,
		GasLimit:         1,
		GasUsed:          1,
		Timestamp:        1,
		ExtraData:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		BaseFeePerGas:    bfpg.SSZBytes(),
		BlockHash:        ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		TransactionsRoot: ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
	}
}

func TestExecutionPayloadHeader_MarshalJSON(t *testing.T) {
	h := &ExecutionPayloadHeader{
		ExecutionPayloadHeader: pbExecutionPayloadHeader(t),
	}
	b, err := json.Marshal(h)
	require.NoError(t, err)
	expected := `{"parent_hash":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","fee_recipient":"0xabcf8e0d4e9587369b2301d0790347320302cc09","state_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","receipts_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","logs_bloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","prev_randao":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","block_number":"1","gas_limit":"1","gas_used":"1","timestamp":"1","extra_data":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","base_fee_per_gas":"452312848583266388373324160190187140051835877600158453279131187530910662656","block_hash":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2","transactions_root":"0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"}`
	require.Equal(t, expected, string(b))
}

var testBuilderBid = `{
    "version":"bellatrix",
	"data":{
		"message":{
			"header":{
				"parent_hash":"0xa0513a503d5bd6e89a144c3268e5b7e9da9dbf63df125a360e3950a7d0d67131",
				"fee_recipient":"0xdfb434922631787e43725c6b926e989875125751",
				"state_root":"0xca3149fa9e37db08d1cd49c9061db1002ef1cd58db2210f2115c8c989b2bdf45",
				"receipts_root":"0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"logs_bloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				"prev_randao":"0xc2fa210081542a87f334b7b14a2da3275e4b281dd77b007bcfcb10e34c42052e",
				"block_number":"1",
				"gas_limit":"10000000",
				"gas_used":"0",
				"timestamp":"4660",
				"extra_data":"0x",
				"base_fee_per_gas":"7",
				"block_hash":"0x10746fa06c248e7eacd4ff8ad8b48a826c227387ee31a6aa5eb4d83ddad34f07",
				"transactions_root":"0x7ffe241ea60187fdb0187bfa22de35d1f9bed7ab061d9401fd47e34a54fbede1"
			},
			"value":"452312848583266388373324160190187140051835877600158453279131187530910662656",
			"pubkey":"0x8645866c95cbc2e08bc77ccad473540eddf4a1f51a2a8edc8d7a673824218f7f68fe565f1ab38dadd5c855b45bbcec95"
		},
		"signature":"0x9183ebc1edf9c3ab2bbd7abdc3b59c6b249d6647b5289a97eea36d9d61c47f12e283f64d928b1e7f5b8a5182b714fa921954678ea28ca574f5f232b2f78cf8900915a2993b396e3471e0655291fec143a300d41408f66478c8208e0f9be851dc"
	}
}`

func TestBuilderBidUnmarshalUint256(t *testing.T) {
	base10 := "452312848583266388373324160190187140051835877600158453279131187530910662656"
	var expectedValue big.Int
	require.NoError(t, expectedValue.UnmarshalText([]byte(base10)))
	r := &ExecHeaderResponse{}
	require.NoError(t, json.Unmarshal([]byte(testBuilderBid), r))
	//require.Equal(t, expectedValue, r.Data.Message.Value)
	marshaled := r.Data.Message.Value.String()
	require.Equal(t, base10, marshaled)
	require.Equal(t, 0, expectedValue.Cmp(r.Data.Message.Value.Int))
}

func TestMathBigUnmarshal(t *testing.T) {
	base10 := "452312848583266388373324160190187140051835877600158453279131187530910662656"
	var expectedValue big.Int
	require.NoError(t, expectedValue.UnmarshalText([]byte(base10)))
	marshaled, err := expectedValue.MarshalText()
	require.NoError(t, err)
	require.Equal(t, base10, string(marshaled))

	var u256 Uint256
	require.NoError(t, u256.UnmarshalText([]byte("452312848583266388373324160190187140051835877600158453279131187530910662656")))
}

func TestUint256Unmarshal(t *testing.T) {
	base10 := "452312848583266388373324160190187140051835877600158453279131187530910662656"
	bi := new(big.Int)
	bi, ok := bi.SetString(base10, 10)
	require.Equal(t, true, ok)
	s := struct {
		BigNumber Uint256 `json:"big_number"`
	}{
		BigNumber: Uint256{Int: bi},
	}
	m, err := json.Marshal(s)
	require.NoError(t, err)
	expected := `{"big_number":"452312848583266388373324160190187140051835877600158453279131187530910662656"}`
	require.Equal(t, expected, string(m))
}

func TestMarshalBlindedBeaconBlockBodyBellatrix(t *testing.T) {
	expected, err := os.ReadFile("testdata/blinded-block.json")
	require.NoError(t, err)
	b := &BlindedBeaconBlockBellatrix{BlindedBeaconBlockBellatrix: &eth.BlindedBeaconBlockBellatrix{
		Slot:          1,
		ProposerIndex: 1,
		ParentRoot:    ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		StateRoot:     ezDecode(t, "0xcf8e0d4e9587369b2301d0790347320302cc0943d5a1884560367e8208d920f2"),
		Body: &eth.BlindedBeaconBlockBodyBellatrix{
			RandaoReveal:           ezDecode(t, "0x1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505cc411d61252fb6cb3fa0017b679f8bb2305b26a285fa2737f175668d0dff91cc1b66ac1fb663c9bc59509846d6ec05345bd908eda73e670af888da41af171505"),
			Eth1Data:               pbEth1Data(),
			Graffiti:               ezDecode(t, "0xdeadbeefc0ffee"),
			ProposerSlashings:      []*eth.ProposerSlashing{pbProposerSlashing(t)},
			AttesterSlashings:      []*eth.AttesterSlashing{pbAttesterSlashing(t)},
			Attestations:           []*eth.Attestation{pbAttestation(t)},
			Deposits:               []*eth.Deposit{pbDeposit(t)},
			VoluntaryExits:         []*eth.SignedVoluntaryExit{pbSignedVoluntaryExit(t)},
			SyncAggregate:          pbSyncAggregate(),
			ExecutionPayloadHeader: pbExecutionPayloadHeader(t),
		},
	}}
	m, err := json.Marshal(b)
	require.NoError(t, err)
	// string error output is easier to deal with
	// -1 end slice index on expected is to get rid of trailing newline
	// if you update this fixture and this test breaks, you probably removed the trailing newline
	require.Equal(t, string(expected[0:len(expected)-1]), string(m))
}

func TestRoundTripUint256(t *testing.T) {
	vs := "4523128485832663883733241601901871400518358776001584532791311875309106626"
	u := stringToUint256(vs)
	sb := u.SSZBytes()
	require.Equal(t, 32, len(sb))
	uu := sszBytesToUint256(sb)
	require.Equal(t, true, bytes.Equal(u.SSZBytes(), uu.SSZBytes()))
	require.Equal(t, vs, uu.String())
}

func TestRoundTripProtoUint256(t *testing.T) {
	h := pbExecutionPayloadHeader(t)
	h.BaseFeePerGas = []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31}
	hm := &ExecutionPayloadHeader{ExecutionPayloadHeader: h}
	m, err := json.Marshal(hm)
	require.NoError(t, err)
	hu := &ExecutionPayloadHeader{}
	require.NoError(t, json.Unmarshal(m, hu))
	hp, err := hu.ToProto()
	require.NoError(t, err)
	require.DeepEqual(t, h.BaseFeePerGas, hp.BaseFeePerGas)
}

func TestExecutionPayloadHeaderRoundtrip(t *testing.T) {
	expected, err := os.ReadFile("testdata/execution-payload.json")
	require.NoError(t, err)
	hu := &ExecutionPayloadHeader{}
	require.NoError(t, json.Unmarshal(expected, hu))
	m, err := json.Marshal(hu)
	require.NoError(t, err)
	require.DeepEqual(t, string(expected[0:len(expected)-1]), string(m))
}