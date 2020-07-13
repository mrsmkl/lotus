package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/build"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"

	"github.com/filecoin-project/lotus/api/apibstore"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/specs-actors/actors/builtin"
	"github.com/filecoin-project/specs-actors/actors/builtin/verifreg"
	"github.com/ipfs/go-hamt-ipld"
	cbor "github.com/ipfs/go-ipld-cbor"

	cbg "github.com/whyrusleeping/cbor-gen"
)

var verifRegCmd = &cli.Command{
	Name:  "verifreg",
	Usage: "Interact with the verified registry actor",
	Flags: []cli.Flag{},
	Subcommands: []*cli.Command{
		verifRegAddVerifierCmd,
		verifRegVerifyClientCmd,
		verifRegListVerifiersCmd,
		verifRegListClientsCmd,
		verifRegCheckClientCmd,
		verifRegCheckVerifierCmd,
		verifRegSetRootCmd,
	},
}

var verifRegAddVerifierCmd = &cli.Command{
	Name:  "add-verifier",
	Usage: "make a given account a verifier",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "dry",
			Usage: "only prints tx param data",
		},
		&cli.StringFlag{
			Name:  "from",
			Usage: "specify your verifier address to send the message from",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must specify two arguments: address and allowance")
		}

		target, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		allowance, err := types.BigFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		params, err := actors.SerializeParams(&verifreg.AddVerifierParams{Address: target, Allowance: allowance})
		if err != nil {
			return err
		}

		if cctx.Bool("dry") {
			fmt.Println(hex.EncodeToString(params))
			return nil
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		froms := cctx.String("from")
		if froms == "" {
			return fmt.Errorf("must specify from address with --from")
		}

		fromk, err := address.NewFromString(froms)
		if err != nil {
			return err
		}

		msg := &types.Message{
			To:       builtin.VerifiedRegistryActorAddr,
			From:     fromk,
			Method:   builtin.MethodsVerifiedRegistry.AddVerifier,
			GasPrice: types.NewInt(1),
			GasLimit: 300000,
			Params:   params,
		}

		smsg, err := api.MpoolPushMessage(ctx, msg)
		if err != nil {
			return err
		}

		fmt.Printf("message sent, now waiting on cid: %s\n", smsg.Cid())

		mwait, err := api.StateWaitMsg(ctx, smsg.Cid(), build.MessageConfidence)
		if err != nil {
			return err
		}

		if mwait.Receipt.ExitCode != 0 {
			return fmt.Errorf("failed to add verifier: %d", mwait.Receipt.ExitCode)
		}

		return nil

	},
}

var verifRegSetRootCmd = &cli.Command{
	Name:  "set-root",
	Usage: "make a key the new root key",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "specify your rootkey address to send the message from",
		},
	},
	Action: func(cctx *cli.Context) error {
		if cctx.Args().Len() != 1 {
			return fmt.Errorf("must specify one argument: address")
		}

		target, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		params, err := actors.SerializeParams(target)
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		froms := cctx.String("from")
		if froms == "" {
			return fmt.Errorf("must specify from address with --from")
		}

		fromk, err := address.NewFromString(froms)
		if err != nil {
			return err
		}

		msg := &types.Message{
			To:   builtin.VerifiedRegistryActorAddr,
			From: fromk,
			// Method:   builtin.MethodsVerifiedRegistry.ReplaceRootKey,
			GasPrice: types.NewInt(1),
			GasLimit: 300000,
			Params:   params,
		}

		smsg, err := api.MpoolPushMessage(ctx, msg)
		if err != nil {
			return err
		}

		fmt.Printf("message sent, now waiting on cid: %s\n", smsg.Cid())

		mwait, err := api.StateWaitMsg(ctx, smsg.Cid(), build.MessageConfidence)
		if err != nil {
			return err
		}

		if mwait.Receipt.ExitCode != 0 {
			return fmt.Errorf("failed to add verifier: %d", mwait.Receipt.ExitCode)
		}

		return nil

	},
}

var verifRegVerifyClientCmd = &cli.Command{
	Name:  "verify-client",
	Usage: "make a given account a verified client",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "from",
			Usage: "specify your verifier address to send the message from",
		},
	},
	Action: func(cctx *cli.Context) error {
		froms := cctx.String("from")
		if froms == "" {
			return fmt.Errorf("must specify from address with --from")
		}

		fromk, err := address.NewFromString(froms)
		if err != nil {
			return err
		}

		if cctx.Args().Len() != 2 {
			return fmt.Errorf("must specify two arguments: address and allowance")
		}

		target, err := address.NewFromString(cctx.Args().Get(0))
		if err != nil {
			return err
		}

		allowance, err := types.BigFromString(cctx.Args().Get(1))
		if err != nil {
			return err
		}

		params, err := actors.SerializeParams(&verifreg.AddVerifiedClientParams{Address: target, Allowance: allowance})
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		msg := &types.Message{
			To:       builtin.VerifiedRegistryActorAddr,
			From:     fromk,
			Method:   builtin.MethodsVerifiedRegistry.AddVerifiedClient,
			GasPrice: types.NewInt(1),
			GasLimit: 300000,
			Params:   params,
		}

		smsg, err := api.MpoolPushMessage(ctx, msg)
		if err != nil {
			return err
		}

		fmt.Printf("message sent, now waiting on cid: %s\n", smsg.Cid())

		mwait, err := api.StateWaitMsg(ctx, smsg.Cid(), build.MessageConfidence)
		if err != nil {
			return err
		}

		if mwait.Receipt.ExitCode != 0 {
			return fmt.Errorf("failed to add verified client: %d", mwait.Receipt.ExitCode)
		}

		return nil
	},
}

var verifRegListVerifiersCmd = &cli.Command{
	Name:  "list-verifiers",
	Usage: "list all verifiers",
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		act, err := api.StateGetActor(ctx, builtin.VerifiedRegistryActorAddr, types.EmptyTSK)
		if err != nil {
			return err
		}

		apibs := apibstore.NewAPIBlockstore(api)
		cst := cbor.NewCborStore(apibs)

		var st verifreg.State
		if err := cst.Get(ctx, act.Head, &st); err != nil {
			return err
		}

		vh, err := hamt.LoadNode(ctx, cst, st.Verifiers)
		if err != nil {
			return err
		}

		if err := vh.ForEach(ctx, func(k string, val interface{}) error {
			addr, err := address.NewFromBytes([]byte(k))
			if err != nil {
				return err
			}

			var dcap verifreg.DataCap

			if err := dcap.UnmarshalCBOR(bytes.NewReader(val.(*cbg.Deferred).Raw)); err != nil {
				return err
			}

			fmt.Printf("%s: %s\n", addr, dcap)

			return nil
		}); err != nil {
			return err
		}

		return nil
	},
}

/*
type hashBits struct {
	b        []byte
	consumed int
}

func mkmask(n int) byte {
	return (1 << uint(n)) - 1
}

func (hb *hashBits) Next(i int) (int, error) {
	if hb.consumed+i > len(hb.b)*8 {
		return 0, fmt.Errorf("sharded directory too deep")
	}
	return hb.next(i), nil
}

func (hb *hashBits) next(i int) int {
	curbi := hb.consumed / 8
	leftb := 8 - (hb.consumed % 8)

	curb := hb.b[curbi]
	if i == leftb {
		out := int(mkmask(i) & curb)
		hb.consumed += i
		return out
	} else if i < leftb {
		a := curb & mkmask(leftb) // mask out the high bits we don't want
		b := a & ^mkmask(leftb-i) // mask out the low bits we don't want
		c := b >> uint(leftb-i)   // shift whats left down
		hb.consumed += i
		return int(c)
	} else {
		out := int(mkmask(leftb) & curb)
		out <<= uint(i - leftb)
		hb.consumed += leftb
		out += hb.next(i - leftb)
		return out
	}
}

func FindRaw(n *hamt.Node, ctx context.Context, k string) ([]byte, error) {
	var ret []byte
	err := getValue(n, ctx, &hashBits{b: n.hash([]byte(k))}, k, func(kv *hamt.KV) error {
		ret = kv.Value.Raw
		return nil
	})
	return ret, err
}

func getValue(n *hamt.Node, ctx context.Context, hv *hashBits, k string, cb func(*KV) error) error {
	idx, err := hv.Next(n.bitWidth)
	if err != nil {
		return ErrMaxDepth
	}

	if n.Bitfield.Bit(idx) == 0 {
		return ErrNotFound
	}

	cindex := byte(n.indexForBitPos(idx))

	c := n.getChild(cindex)
	if c.isShard() {
		chnd, err := c.loadChild(ctx, n.store, n.bitWidth, n.hash)
		if err != nil {
			return err
		}

		return chnd.getValue(ctx, hv, k, cb)
	}

	for _, kv := range c.KVs {
		if string(kv.Key) == k {
			return cb(kv)
		}
	}

	return ErrNotFound
}
*/

var verifRegListClientsCmd = &cli.Command{
	Name:  "list-clients",
	Usage: "list all verified clients",
	Action: func(cctx *cli.Context) error {
		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		act, err := api.StateGetActor(ctx, builtin.VerifiedRegistryActorAddr, types.EmptyTSK)
		if err != nil {
			return err
		}

		apibs := apibstore.NewAPIBlockstore(api)
		cst := cbor.NewCborStore(apibs)

		var st verifreg.State
		if err := cst.Get(ctx, act.Head, &st); err != nil {
			return err
		}

		vh, err := hamt.LoadNode(ctx, cst, st.VerifiedClients)
		if err != nil {
			return err
		}
		log.Infof("Loaded node %v %v", len(vh.Pointers), vh.Bitfield)

		// var dcap verifreg.DataCap
		addr, _ := address.NewFromString("t1xdzis7pauuihealpvdfz7kj5pcjizil7sbtblni")
		log.Infof("fooq %v", addr.Bytes())
		if _, err := vh.FindRaw(ctx, string(addr.Bytes())); err != nil {
			log.Warnf("what %w", err)
		}

		if err := vh.ForEach(ctx, func(k string, val interface{}) error {
			addr, err := address.NewFromBytes([]byte(k))
			log.Infof("fooq %v %v", addr.Bytes(), []byte(k))
			if err != nil {
				return err
			}

			var dcap verifreg.DataCap

			if err := dcap.UnmarshalCBOR(bytes.NewReader(val.(*cbg.Deferred).Raw)); err != nil {
				return err
			}

			fmt.Printf("%s: %s\n", addr, dcap)

			return nil
		}); err != nil {
			return err
		}

		return nil
	},
}

var verifRegCheckClientCmd = &cli.Command{
	Name:  "check-client",
	Usage: "check verified client remaining bytes",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must specify client address to check")
		}

		caddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		dcap, err := api.StateVerifiedClientStatus(ctx, caddr, types.EmptyTSK)
		if err != nil {
			return err
		}
		if dcap == nil {
			return xerrors.Errorf("client %s is not a verified client", err)
		}

		fmt.Println(*dcap)

		return nil
	},
}

var verifRegCheckVerifierCmd = &cli.Command{
	Name:  "check-verifier",
	Usage: "check verifiers remaining bytes",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must specify verifier address to check")
		}

		vaddr, err := address.NewFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		api, closer, err := lcli.GetFullNodeAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()
		ctx := lcli.ReqContext(cctx)

		act, err := api.StateGetActor(ctx, builtin.VerifiedRegistryActorAddr, types.EmptyTSK)
		if err != nil {
			return err
		}

		apibs := apibstore.NewAPIBlockstore(api)
		cst := cbor.NewCborStore(apibs)

		var st verifreg.State
		if err := cst.Get(ctx, act.Head, &st); err != nil {
			return err
		}

		vh, err := hamt.LoadNode(ctx, cst, st.Verifiers, hamt.UseTreeBitWidth(5))
		if err != nil {
			return err
		}

		fmt.Printf("loaded node %v\n", vh)

		var dcap verifreg.DataCap
		if err := vh.Find(ctx, string(vaddr.Bytes()), &dcap); err != nil {
			return err
		}

		fmt.Println(dcap)

		return nil
	},
}
