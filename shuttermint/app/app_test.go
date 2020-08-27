package app

import (
	"crypto/ecdsa"
	"testing"
	"unicode/utf8"

	"github.com/brainbot-com/shutter/shuttermint/shmsg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/kv"
)

func TestNewShutterApp(t *testing.T) {
	app := NewShutterApp()
	require.Equal(t, len(app.Configs), 1, "Configs should contain exactly one guard element")
	require.Equal(t, app.Configs[0], &BatchConfig{}, "Bad guard element")
}

func TestGetBatch(t *testing.T) {
	app := NewShutterApp()

	err := app.addConfig(BatchConfig{StartBatchIndex: 100, Threshhold: 1})
	require.Nil(t, err)

	err = app.addConfig(BatchConfig{StartBatchIndex: 200, Threshhold: 2})
	require.Nil(t, err)

	err = app.addConfig(BatchConfig{StartBatchIndex: 300, Threshhold: 3})
	require.Nil(t, err)

	require.Equal(t, app.getBatch(0).Config.Threshhold, uint32(0))
	require.Equal(t, app.getBatch(99).Config.Threshhold, uint32(0))
	require.Equal(t, app.getBatch(100).Config.Threshhold, uint32(1))
	require.Equal(t, app.getBatch(101).Config.Threshhold, uint32(1))
	require.Equal(t, app.getBatch(199).Config.Threshhold, uint32(1))
	require.Equal(t, app.getBatch(200).Config.Threshhold, uint32(2))
	require.Equal(t, app.getBatch(1000).Config.Threshhold, uint32(3))
}

func TestAddConfig(t *testing.T) {
	app := NewShutterApp()

	err := app.addConfig(BatchConfig{StartBatchIndex: 100, Threshhold: 1})
	require.Nil(t, err)

	err = app.addConfig(BatchConfig{StartBatchIndex: 50, Threshhold: 1})
	require.NotNil(t, err, "Expected error, StartBatchIndex must increase")

	err = app.addConfig(BatchConfig{StartBatchIndex: 100, Threshhold: 1})
	require.NotNil(t, err, "Expected error, StartBatchIndex must increase")
}

func TestKeyGeneration(t *testing.T) {
	app := NewShutterApp()
	keypers := addresses[:3]

	var keys [3]*ecdsa.PrivateKey
	for i := 0; i < len(keys); i++ {
		k, err := crypto.GenerateKey()
		if err != nil {
			t.Fatalf("Could not generate key: %s", err)
		}
		keys[i] = k
	}

	err := app.addConfig(BatchConfig{StartBatchIndex: 100, Threshhold: 2, Keypers: keypers})
	require.Nil(t, err)
	res1 := app.deliverPublicKeyCommitment(
		&shmsg.PublicKeyCommitment{
			BatchIndex: 200,
			Commitment: crypto.FromECDSAPub(&keys[0].PublicKey)},
		keypers[0])
	require.Equal(
		t,
		types.ResponseDeliverTx{Code: 0, Events: []types.Event(nil)},
		res1)

	res2 := app.deliverPublicKeyCommitment(
		&shmsg.PublicKeyCommitment{
			BatchIndex: 200,
			Commitment: crypto.FromECDSAPub(&keys[1].PublicKey)},
		keypers[1])
	// We've reached the threshold, there should be an event of Type "shutter.pubkey-generated"
	require.Equal(
		t,
		types.ResponseDeliverTx{
			Code: 0,
			Events: []types.Event{
				{
					Type: "shutter.pubkey-generated",
					Attributes: []kv.Pair{
						{
							Key:   []byte("BatchIndex"),
							Value: []byte("200"),
						},
						{
							Key:   []byte("Pubkey"),
							Value: []byte(encodePubkeyForEvent(&keys[1].PublicKey)),
						},
					},
				}}},
		res2)
	res3 := app.deliverPublicKeyCommitment(
		&shmsg.PublicKeyCommitment{
			BatchIndex: 200,
			Commitment: crypto.FromECDSAPub(&keys[2].PublicKey)},
		keypers[2])
	require.Equal(
		t,
		types.ResponseDeliverTx{Code: 0, Events: []types.Event(nil)},
		res3)

}

func TestEncodePubkeyForEvent(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.Nil(t, err, "Could not generate key")
	encoded := encodePubkeyForEvent(&key.PublicKey)
	t.Logf("Encoded: %s", encoded)
	require.True(t, utf8.ValidString(encoded))

	decoded, err := decodePubkeyFromEvent(encoded)
	require.Nil(t, err, "could not decode pubkey")
	t.Logf("Decoded: %+v", decoded)
	require.Equal(t, key.PublicKey, *decoded)
}