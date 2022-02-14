package ipfs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	cid "github.com/ipfs/go-cid"
	// shell "github.com/ipfs/go-ipfs-api"
	options "github.com/ipfs/go-ipfs-api/options"
	config "github.com/ipfs/go-ipfs-config"
	files "github.com/ipfs/go-ipfs-files"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreapi"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/plugin/loader"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	icore "github.com/ipfs/interface-go-ipfs-core"
	icorepath "github.com/ipfs/interface-go-ipfs-core/path"
	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"

	blockstore "github.com/ipfs/go-ipfs-blockstore"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	bsadapter "github.com/ipld/go-ipld-prime/storage/bsadapter"
)

// var sh *shell.Shell

type IpfsContext struct {
	value      uint
	node       icore.CoreAPI
	blockstore blockstore.GCBlockstore
	ctx        context.Context
}

// type ExtendedCoreApiStruct struct {
// 	// core       *coreapi.CoreAPI
// 	core       *icore.CoreAPI
// 	blockstore blockstore.GCBlockstore
// }

// type ExtendedCoreApi interface {
// 	icore.CoreAPI
// 	BlockStore() blockstore.GCBlockstore
// }

// func (s *ExtendedCoreApiStruct) BlockStore() blockstore.GCBlockstore {
// 	return s.blockstore
// }

func NewIpfs() *IpfsContext {

	// r := strings.NewReader(`{"hey":"it works!","yes": true}`)
	// fmt.Println("--DagPutWithOpts-r--", r)

	// np := basicnode.Prototype.Any // Pick a stle for the in-memory data.
	// nb := np.NewBuilder()         // Create a builder.
	// dagjson.Decode(nb, r)         // Hand the builder to decoding -- decoding will fill it in!

	// fmt.Println("--DagPutWithOpts-nb--", nb)

	// n := nb.Build() // Call 'Build' to get the resulting Node.  (It's immutable!)

	// fmt.Println("--DagPutWithOpts-n--", n)

	// fmt.Printf("the data decoded was a %s kind\n", n.Kind())
	// fmt.Printf("the length of the node is %d\n", n.Length())

	// flag.Parse()
	fmt.Println("-- Getting an IPFS node running -- ")

	// ctx, cancel := context.WithCancel(context.Background())
	ctx, _ := context.WithCancel(context.Background())
	// defer cancel()

	// Spawn a node using the default path (~/.ipfs), assuming that a repo exists there already
	fmt.Println("Spawning node on default repo")
	ipfs, blockstore, err := spawnEphemeral(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to spawnDefault node: %s", err))
	}

	// // Spawn a node using a temporary path, creating a temporary repo for the run
	// fmt.Println("--Spawning node on a temporary repo")
	// ipfs, err := spawnEphemeral(ctx)
	// if err != nil {
	// 	panic(fmt.Errorf("--failed to spawn ephemeral node: %s", err))
	// }

	fmt.Println("--IPFS node is running")

	fmt.Println("\n-- Going to connect to a few nodes in the Network as bootstrappers --")

	bootstrapNodes := []string{
		// IPFS Bootstrapper nodes.
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmbLHAnMoJPWSCR5Zhtx6BHJX9KiKNN6tpvbUcqanj75Nb",
		"/dnsaddr/bootstrap.libp2p.io/p2p/QmcZf59bWwK5XFi76CZX8cbJ4BhTzzA3gU1ZjYZcYW3dwt",

		"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/104.131.131.82/udp/4001/quic/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
		"/ip4/46.101.103.114/tcp/4001/p2p/QmYfXRuVWMWFRJxUSFPHtScTNR9CU2samRsTK15VFJPpvh",

		// IPFS Cluster Pinning nodes
		"/ip4/138.201.67.219/tcp/4001/p2p/QmUd6zHcbkbcs7SMxwLs48qZVX3vpcM8errYS7xEczwRMA",
		"/ip4/138.201.67.219/udp/4001/quic/p2p/QmUd6zHcbkbcs7SMxwLs48qZVX3vpcM8errYS7xEczwRMA",
		"/ip4/138.201.67.220/tcp/4001/p2p/QmNSYxZAiJHeLdkBg38roksAR9So7Y5eojks1yjEcUtZ7i",
		"/ip4/138.201.67.220/udp/4001/quic/p2p/QmNSYxZAiJHeLdkBg38roksAR9So7Y5eojks1yjEcUtZ7i",
		"/ip4/138.201.68.74/tcp/4001/p2p/QmdnXwLrC8p1ueiq2Qya8joNvk3TVVDAut7PrikmZwubtR",
		"/ip4/138.201.68.74/udp/4001/quic/p2p/QmdnXwLrC8p1ueiq2Qya8joNvk3TVVDAut7PrikmZwubtR",
		"/ip4/94.130.135.167/tcp/4001/p2p/QmUEMvxS2e7iDrereVYc5SWPauXPyNwxcy9BXZrC1QTcHE",
		"/ip4/94.130.135.167/udp/4001/quic/p2p/QmUEMvxS2e7iDrereVYc5SWPauXPyNwxcy9BXZrC1QTcHE",

		// your
		"/ip4/127.0.0.1/tcp/4001/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip4/127.0.0.1/udp/4001/quic/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip6/::1/tcp/4001/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip6/::1/udp/4001/quic/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
	}

	go func() {
		err := connectToPeers(ctx, ipfs, bootstrapNodes)
		if err != nil {
			log.Printf("failed connect to peers: %s", err)
		}
	}()

	return &IpfsContext{
		value:      3,
		node:       ipfs,
		ctx:        ctx,
		blockstore: blockstore,
	}
}

func (s *IpfsContext) GetValue() uint                      { return s.value }
func (s *IpfsContext) Node() icore.CoreAPI                 { return s.node }
func (s *IpfsContext) Ctx() context.Context                { return s.ctx }
func (s *IpfsContext) Blockstore() blockstore.GCBlockstore { return s.blockstore }

func (s *IpfsContext) SaveFile(content []byte) ([]byte, error) {
	someFile := files.NewBytesFile(content)
	cidFile, err := s.Node().Unixfs().Add(s.Ctx(), someFile)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Added file to IPFS with CID %s\n", cidFile.String())
	return cidFile.Cid().Bytes(), nil
}

func (s *IpfsContext) LoadFile(cidBytes []byte) ([]byte, error) {
	cidValue, err := cid.Parse(cidBytes)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Fetching a file from the network with CID %s\n", cidValue.String())
	cidPath := icorepath.New(cidValue.String())
	rootNode, err := s.Node().Unixfs().Get(s.Ctx(), cidPath)
	if err != nil {
		return nil, err
	}
	rootFile := files.ToFile(rootNode)
	buf := new(bytes.Buffer)
	buf.ReadFrom(rootFile)
	result := buf.Bytes()
	// newStr := buf.String()
	// fmt.Printf("data %s", newStr)
	return result, nil
}

func (s *IpfsContext) IpldPut(content []byte) ([]byte, error) {
	return s.DagPutWithOpts(content, options.Dag.InputCodec("json"), options.Dag.StoreCodec("cbor"))
}

func (s *IpfsContext) IpldGet(cidBytes []byte, key []byte) ([]byte, error) {
	cidValue, err := cid.Parse(cidBytes)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Fetching a file from the network with CID %s\n", cidValue.String())

	fmt.Println("--IpldGet-key--", string(key))

	keyValue := string(key)

	lsys := cidlink.DefaultLinkSystem()
	store := &bsadapter.Adapter{
		Wrapped: s.Blockstore(),
	}

	// return store.Get(s.Ctx(), keyValue)

	lsys.SetReadStorage(store)
	// lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
	// 	Version:  1,
	// 	Codec:    0x71, // "dag-cbor"
	// 	MhType:   0x12, // sha2-256
	// 	MhLength: 32,   //
	// }}.BuildLink(cidBytes)

	lp := cidlink.Link{Cid: cidValue}
	np := basicnode.Prototype.Any // Pick a stle for the in-memory data.

	node, err := lsys.Load(
		linking.LinkContext{}, // The zero value is fine.  Configure it it you want cancellability or other features.
		lp,                    // The LinkPrototype says what codec and hashing to use.
		np,                    // And here's our data.
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("we loaded a %s with %d entries\n", node.Kind(), node.Length(), keyValue)

	node2, err := node.LookupByString(keyValue)

	fmt.Printf("we loaded a %s with %d entries\n", node2.Kind(), node2.Length())

	if node2.IsAbsent() {
		return nil, nil
	}
	if node2.IsNull() {
		return nil, nil
	}

	switch node2.Kind() {
	case datamodel.Kind_Bool:
		value, err := node2.AsBool()
		if err != nil {
			return nil, err
		}
		v := 0
		if value {
			v = 1
		}
		return new(big.Int).SetInt64(int64(v)).FillBytes(make([]byte, 32)), nil
	case datamodel.Kind_Int:
		value, err := node2.AsInt()
		if err != nil {
			return nil, err
		}
		return new(big.Int).SetInt64(int64(value)).FillBytes(make([]byte, 32)), nil
	case datamodel.Kind_Float:
		return nil, fmt.Errorf("float not supported")
	case datamodel.Kind_String:
		value, err := node2.AsString()
		if err != nil {
			return nil, err
		}
		return []byte(value), nil
	case datamodel.Kind_Bytes:
		return node2.AsBytes()
	case datamodel.Kind_List:
		return nil, fmt.Errorf("list not supported")
	case datamodel.Kind_Map:
		return nil, fmt.Errorf("map not supported")
	case datamodel.Kind_Link:
		return nil, fmt.Errorf("link not supported")
	default:
		return nil, nil
	}
	return nil, nil
}

func (s *IpfsContext) DagPutWithOpts(data []byte, opts ...options.DagPutOption) ([]byte, error) {
	r := bytes.NewReader(data)

	np := basicnode.Prototype.Any // Pick a stle for the in-memory data.
	nb := np.NewBuilder()         // Create a builder.
	dagjson.Decode(nb, r)         // Hand the builder to decoding -- decoding will fill it in!
	// raw.Decode(nb, r)

	n := nb.Build() // Call 'Build' to get the resulting Node.  (It's immutable!)

	fmt.Printf("the data decoded was a %s kind\n", n.Kind())
	fmt.Printf("the length of the node is %d\n", n.Length())

	lsys := cidlink.DefaultLinkSystem()
	store := &bsadapter.Adapter{
		Wrapped: s.Blockstore(),
	}

	lsys.SetWriteStorage(store)
	// https://github.com/multiformats/multicodec/
	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,
		Codec:    0x71, // "dag-cbor"
		MhType:   0x12, // sha2-256
		MhLength: 32,   //
	}}

	lnk, err := lsys.Store(
		linking.LinkContext{}, // The zero value is fine.  Configure it it you want cancellability or other features.
		lp,                    // The LinkPrototype says what codec and hashing to use.
		n,                     // And here's our data.
	)
	if err != nil {
		return nil, err
	}

	// That's it!  We got a link.
	fmt.Printf("link: %s\n", lnk)
	fmt.Printf("concrete type: `%T`\n", lnk)
	fmt.Println("lnk.String()", lnk.String(), lnk.Binary())

	return []byte(lnk.String()), nil
}

// var flagExp = flag.Bool("experimental", false, "enable experimental features")

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func createTempRepo() (string, error) {
	repoPath, err := ioutil.TempDir("", "ipfs-shell")
	if err != nil {
		return "", fmt.Errorf("failed to get temp dir: %s", err)
	}

	// Create a config with default options and a 2048 bit key
	cfg, err := config.Init(ioutil.Discard, 2048)
	if err != nil {
		return "", err
	}

	// When creating the repository, you can define custom settings on the repository, such as enabling experimental
	// features (See experimental-features.md) or customizing the gateway endpoint.
	// To do such things, you should modify the variable `cfg`. For example:
	// if *flagExp {
	// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#ipfs-filestore
	cfg.Experimental.FilestoreEnabled = true
	// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#ipfs-urlstore
	cfg.Experimental.UrlstoreEnabled = true
	// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#ipfs-p2p
	cfg.Experimental.Libp2pStreamMounting = true
	// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#p2p-http-proxy
	cfg.Experimental.P2pHttpProxy = true
	// https://github.com/ipfs/go-ipfs/blob/master/docs/experimental-features.md#strategic-providing
	cfg.Experimental.StrategicProviding = true
	// }

	// Create the repo with the config
	err = fsrepo.Init(repoPath, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to init ephemeral node: %s", err)
	}

	return repoPath, nil
}

func createNode(ctx context.Context, repoPath string) (icore.CoreAPI, blockstore.GCBlockstore, error) {
	fmt.Println("--Open repo", repoPath)
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, nil, err
	}

	// Construct the node

	nodeOptions := &core.BuildCfg{
		Online:  true,
		Routing: libp2p.DHTOption, // This option sets the node to be a full DHT node (both fetching and storing DHT Records)
		// Routing: libp2p.DHTClientOption, // This option sets the node to be a client DHT node (only fetching records)
		Repo: repo,
	}

	node, err := core.NewNode(ctx, nodeOptions)
	if err != nil {
		return nil, nil, err
	}

	// fmt.Println("-----nodeOptions---", node.Blockstore)

	fmt.Println("-----node IsOnline", node.IsOnline)
	conf, err := node.Repo.Config()
	fmt.Println("-----node conf", conf, err)
	peers, err2 := conf.BootstrapPeers()
	fmt.Println("---cfg.BootstrapPeers()", peers, err2)

	// Attach the Core API to the constructed node
	newnode, err := coreapi.NewCoreAPI(node)
	if err != nil {
		return nil, nil, err
	}
	return newnode, node.Blockstore, nil
}

// Spawns a node on the default repo location, if the repo exists
func spawnDefault(ctx context.Context) (icore.CoreAPI, blockstore.GCBlockstore, error) {
	defaultPath, err := config.PathRoot()
	if err != nil {
		// shouldn't be possible
		return nil, nil, err
	}

	if err := setupPlugins(defaultPath); err != nil {
		return nil, nil, err

	}

	return createNode(ctx, defaultPath)
}

// Spawns a node to be used just for this run (i.e. creates a tmp repo)
func spawnEphemeral(ctx context.Context) (icore.CoreAPI, blockstore.GCBlockstore, error) {
	if err := setupPlugins(""); err != nil {
		return nil, nil, err
	}

	// Create a Temporary Repo
	repoPath, err := createTempRepo()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create temp repo: %s", err)
	}

	// Spawning an ephemeral IPFS node
	return createNode(ctx, repoPath)
}

func connectToPeers(ctx context.Context, ipfs icore.CoreAPI, peers []string) error {
	var wg sync.WaitGroup
	peerInfos := make(map[peer.ID]*peer.AddrInfo, len(peers))
	for _, addrStr := range peers {
		addr, err := ma.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		pii, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		pi, ok := peerInfos[pii.ID]
		if !ok {
			pi = &peer.AddrInfo{ID: pii.ID}
			peerInfos[pi.ID] = pi
		}
		pi.Addrs = append(pi.Addrs, pii.Addrs...)
	}

	wg.Add(len(peerInfos))
	for _, peerInfo := range peerInfos {
		go func(peerInfo *peer.AddrInfo) {
			defer wg.Done()
			err := ipfs.Swarm().Connect(ctx, *peerInfo)
			if err != nil {
				log.Printf("failed to connect to %s: %s", peerInfo.ID, err)
			}
		}(peerInfo)
	}
	wg.Wait()
	return nil
}

func getUnixfsFile(path string) (files.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	st, err := file.Stat()
	if err != nil {
		return nil, err
	}

	f, err := files.NewReaderPathFile(path, file, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func getUnixfsNode(path string) (files.Node, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	f, err := files.NewSerialFile(path, false, st)
	if err != nil {
		return nil, err
	}

	return f, nil
}
