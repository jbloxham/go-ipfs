package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/ipfs/go-ipfs/core/coreapi/interface"

	cmds "gx/ipfs/Qma6uuSyjkecGhMFFLfzyJDPyoDtNJSHJNweDccZhaWkgU/go-ipfs-cmds"
	coreiface "github.com/ipfs/go-ipfs/core/coreapi/interface"
	files "gx/ipfs/QmZMWMvWMVKCbHetJ4RgndbuEF1io2UpUxwQwtNjtYPzSC/go-ipfs-files"
	options "github.com/ipfs/go-ipfs/core/coreapi/interface/options"
	"gx/ipfs/Qmde5VP1qUkyQXKCfmEUA7bP64V2HAptbJ7phuPp7jXWwg/go-ipfs-cmdkit"
)

const indexOutChanSize = 8

type SearchIndex map[string]map[string]float64;

var IndexCmd = &cmds.Command{
	Helptext: cmdkit.HelpText{
		Tagline: "Index a file for searching on IPFS.",
		ShortDescription: `
'ipfs index' will index yo files.
		`,
	},
	Arguments: []cmdkit.Argument{
		cmdkit.StringArg("ipfs-path", true, true, "The path to the IPFS object to be indexed.").EnableStdin(),
	},
	Options: []cmdkit.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env)
		if err != nil {
			return err
		}

		index, err := index(req.Context, api, req.Arguments)
		if err != nil {
			return err
		}

		return write(res, req.Context, api, index)
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			outChan := make(chan interface{})
			req := res.Request()

			progressBar := func(wait chan struct{}) {
				defer close(wait)

			LOOP:
				for {
					select {
					case out, ok := <-outChan:
						if !ok {
							break LOOP
						}
						output := out.(*coreiface.AddEvent)
						if len(output.Hash) > 0 {
							fmt.Fprintf(os.Stdout, "added %s %s\n", output.Hash, output.Name)
						}
					case <-req.Context.Done():
						// don't set or print error here, that happens in the goroutine below
						return
					}
				}
			}

			if e := res.Error(); e != nil {
				close(outChan)
				return e
			}

			wait := make(chan struct{})
			go progressBar(wait)

			defer func() { <-wait }()
			defer close(outChan)

			for {
				v, err := res.Next()
				if err != nil {
					if err == io.EOF {
						return nil
					}

					return err
				}

				select {
				case outChan <- v:
				case <-req.Context.Done():
					return req.Context.Err()
				}
			}
		},
	},
}

func index(ctx context.Context, api iface.CoreAPI, paths []string) (SearchIndex, error) {
	index := make(SearchIndex)
	for _, p := range paths {
		fpath, err := iface.ParsePath(p)
		if err != nil {
			return nil, err
		}

		file, err := api.Unixfs().Get(ctx, fpath)
		if err != nil {
			return nil, err
		}

		if file.IsDirectory() {
			return nil, iface.ErrIsDir
		}

		var data []byte
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				return nil, err
			}
			data = append(data, buf[:n]...)
			if err == io.EOF {
				break
			}
		}

		keywords := strings.Fields(string(data))

		for _, keyword := range keywords {
			_, prs := index[keyword]
			if prs {
				index[keyword][p] += 1
			} else {
				index[keyword] = map[string]float64{p:1}
			}
		}
	}

	return index, nil
}

// Read the old index file as a kv store
func read(ctx context.Context, api iface.CoreAPI) SearchIndex {
	return nil
}

// Write the bytes of the new index file
func write(res cmds.ResponseEmitter, ctx context.Context, api iface.CoreAPI, index SearchIndex) error {
	jsonIndex, err := json.Marshal(index)
	if err != nil {
		return err
	}

	file := files.NewReaderFile("", "", ioutil.NopCloser(strings.NewReader(string(jsonIndex))), nil)

	events := make(chan interface{}, indexOutChanSize)
	opts := []options.UnixfsAddOption{
		options.Unixfs.Events(events),
	}

	errCh := make(chan error)
	go func() {
		var err error
		defer func() { errCh <- err }()
		defer close(events)
		_, err = api.Unixfs().Add(ctx, file, opts...)
	}()

	err = res.Emit(events)
	if err != nil {
		return err
	}

	return <-errCh
}
