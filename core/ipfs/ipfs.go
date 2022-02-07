package ipfs

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	cid "github.com/ipfs/go-cid"
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
)

type IpfsContext struct {
	value uint
	node  icore.CoreAPI
	ctx   context.Context
}

func NewIpfs() *IpfsContext {

	// flag.Parse()
	fmt.Println("-- Getting an IPFS node running -- ")

	// ctx, cancel := context.WithCancel(context.Background())
	ctx, _ := context.WithCancel(context.Background())
	// defer cancel()

	// Spawn a node using the default path (~/.ipfs), assuming that a repo exists there already
	fmt.Println("Spawning node on default repo")
	ipfs, err := spawnEphemeral(ctx)
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

		// mine
		"/ip4/127.0.0.1/tcp/4001/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip4/127.0.0.1/udp/4001/quic/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip4/192.168.0.106/tcp/4001/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip4/192.168.0.106/udp/4001/quic/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
		"/ip4/85.203.44.147/udp/4001/quic/p2p/Qmf4ypjr6LeTnZroCzNUedKyBjDnDZMtGy2C7xUPrsR7rx",
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
		value: 3,
		node:  ipfs,
		ctx:   ctx,
	}
}

func (s *IpfsContext) GetValue() uint       { return s.value }
func (s *IpfsContext) Node() icore.CoreAPI  { return s.node }
func (s *IpfsContext) Ctx() context.Context { return s.ctx }

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

func createNode(ctx context.Context, repoPath string) (icore.CoreAPI, error) {
	fmt.Println("--Open repo", repoPath)
	// Open the repo
	repo, err := fsrepo.Open(repoPath)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	fmt.Println("-----node IsOnline", node.IsOnline)
	conf, err := node.Repo.Config()
	fmt.Println("-----node conf", conf, err)
	peers, err2 := conf.BootstrapPeers()
	fmt.Println("---cfg.BootstrapPeers()", peers, err2)

	// Attach the Core API to the constructed node
	return coreapi.NewCoreAPI(node)
}

// Spawns a node on the default repo location, if the repo exists
func spawnDefault(ctx context.Context) (icore.CoreAPI, error) {
	defaultPath, err := config.PathRoot()
	if err != nil {
		// shouldn't be possible
		return nil, err
	}

	if err := setupPlugins(defaultPath); err != nil {
		return nil, err

	}

	return createNode(ctx, defaultPath)
}

// Spawns a node to be used just for this run (i.e. creates a tmp repo)
func spawnEphemeral(ctx context.Context) (icore.CoreAPI, error) {
	if err := setupPlugins(""); err != nil {
		return nil, err
	}

	// Create a Temporary Repo
	repoPath, err := createTempRepo()
	if err != nil {
		return nil, fmt.Errorf("failed to create temp repo: %s", err)
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
