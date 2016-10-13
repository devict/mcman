package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/public/css/mcman.css": {
		local:   "public/css/mcman.css",
		size:    1539,
		modtime: 1476371015,
		compressed: `
H4sIAAAJbogA/4SUT5OjLBDG734Kqt5Lpio4ajLJBC/vF9jTnvaIgkoN0BbgJrPWfPfFP2OiMdmqsQZ4
noYf3U2yxjnQbU0ZE7ok7/Ul/QrCYRXbJs+5tdvvOTcGzDQ7U6N9zDS3PAfNqPlsgxwkGILOlXA8zcAw
brChTDSWoL0/wvGLw7aiDM4ERSiuL/1nyoxuoi0a/8Lk5R6mzWj+URpoNCNdwCZ536L4fb9Fh7dbe896
b44Sb4y679Y8XuXenuz85vFxi5KZ/3rVZcThMNIkSdxFvL6iH1w36KdriiIIVa6oP87Quua3dHjI2H9F
wXanKFXUlELjDPxpisRcpfjMsw/hcAHaZ04BuKqrF9VOUCmo5SwNKi7KypEk7ALgNzeF9PmtBGNcTxs4
Q7UVToAmg9/n+c2mWMGfR1qAlX2kPVj+Wlw1hJrrdgSM955vciifnrZvCH+RUhPTea6yg7KUvD0L5iqy
65pn3KUfM2FrST9JJiH/SGsYUWhmQTa+9xzUJEr7Lf3/b7cGzZcnoDCjaxU5Ho+LUwaUJLqiJH447/I4
6uRgHSjuXtnAdPSjlcpQKdfLMgnLmkzC2traVUkhjHU4r4Rk7QyhAKNIP5LU8V8bfKgvLzck6/oVaF1/
Ii3wwstQi3soA86HbfZvjJd3RHNxgTMXH62vg/wjU+MW+CkVfoqFH3HhCex/xZmgaKPoBQ8NuD+Gp9OJ
q5c2zBvrfyf6l4R3t29J8sI/pW99fErzbv4K/gYAAP//z9yGmgMGAAA=
`,
	},

	"/public/css/pure.css": {
		local:   "public/css/pure.css",
		size:    27204,
		modtime: 1475586403,
		compressed: `
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
`,
	},

	"/public/js/B.js": {
		local:   "public/js/B.js",
		size:    7237,
		modtime: 1475586403,
		compressed: `
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
`,
	},

	"/public/js/mcman.js": {
		local:   "public/js/mcman.js",
		size:    953,
		modtime: 1475586403,
		compressed: `
H4sIAAAJbogA/5xTX2vbMBB/dj+FyB4kg+NkD3tZCGN/zDJoMxhhHZQShHxJRBXJs+SkdPi7705unazO
0wIxifS7+/3xndg0VgXtrDhqW7pjVjrV7MGG9M9V8knwNz64am20feBpjiiujFYPPOvLCJccZM3cD/jN
5szCkf26uV6EUOFBAz6IdIYQusYGxskSUWflylnvDOTGbUXYaZ/X4Cs8gxU8hnTWnoorsGJUOR9GGRtN
ZKUnh7cT0of/Q93AiceDLSNtSw9Sh5aaOfmhHzzNEJncflt++X67/rz4uPxarIufxXI1F9xZV2sMQJI+
tZN2C5xpy7p80g98eP2eo2b9BBzJXpyx4LZbAwtEPzmEmy6pu/t84+pCql2upDGCdCQvuliuGvSzHytp
x6GW1iN2/6y27yzAxFb4AZMrI72/1j7kHaHgVVPDmNqNdz05j9kkSYvPmMtQ6Q1WdBr1RlD53fT+rLui
Ptp6DAjfA087CR7CSu/BNUG89pu9m04ja8vAeGARPwwlIvBLjBfMdGQEopC6U5zEIfCRD3wp4/z/2TpP
5Fkgdf5HwtkynMYZYv2r8jiE3fjksiyLA44PkYOFWlwYwqwXjoVtGrcii6vRL+fs6m8AAAD//4RPaIS5
AwAA
`,
	},

	"/tmpls/_layout.go.html": {
		local:   "tmpls/_layout.go.html",
		size:    1365,
		modtime: 1476370294,
		compressed: `
H4sIAAAJbogA/7RUwW4bIRA9218x2eZYlqSRqipa+9A0h0i12kMuVZQDBnZ3EhZWwLpyI/97WSB2EjuW
euheYOHNzHvDzFQn335c3f76eQ2t79R8WqVlUrWSifkUwld59ErOn55gqQx/hCL+F1DCZiPk6ubqFhao
Jbes9gEktQgXFU1W08mkUqgfwUo1K5xfK+laKX0BrZX1rKD9sFTIKXcubK0sw6b4J6uOd0xns+qEkDus
QXkJN9fw5T4rOOar9b6/pHQ9YLlmrTGsR1dy00U69Kz8XJ7RxqJwxErXG+1wJYlRgqAkHe4i3wXhWN8T
sqPR+MxiPPhfVF5yIOQVj0lF0zNOqqUR63EVuAKumHOzIiaO/Las76WFMQRpCkARbqQe0iO8gEfAQM4h
bzpBzslFhO3jth7CHcvqPhR7EDLSQ90AH5w3HVlapkUxX1yRBdMVZe96SNy9aZqxEkfOeT+v3DNmyWz4
pe7gUfZd0cA8Kt1ujkv+RC4gBd9JPKgeXog0Fv8Y7ZnKpnzkHqS62tjuOU2D2k+PQueLVDjpC/0V7BoJ
p/gRTldwOYNyMUbbbKKXsW323aCXIcw2jcHJ6ar8bjjzaHQwPfAwY5kW84xkS6liT7OQOIVvCKWGTyLo
oA7nNa+hAd6rwFx6DDUZizXL3k2deBiHzjb8bYsOehbSEVZtPAhZh0kkyulraiHqMw3HLfYenOW7KfLg
6NfywcXKiNdHgGnavAXT1F6h3eL0/BsAAP//RSN5uFUFAAA=
`,
	},

	"/tmpls/index.go.html": {
		local:   "tmpls/index.go.html",
		size:    537,
		modtime: 1476371135,
		compressed: `
H4sIAAAJbogA/3RRwW6rMBC85ytWK44PENc84NRj1V7aD3DiLVhyTWUb1CjKv3fXIXWaNjdGMzszHo5H
0PRmHAHuJn1AOJ02AK02C+ytCqHDj9lTOcCeXCSPPbPMj80Pei4bhBAPljqM9BlLZc3gtuvRf+yfnZWM
10A+tPXYrDZR7SzdOkH6SBSC0R1O6bic5XgtILcjKX1Bgn0Gib74Yi+x8KTeqa3jeFf1OA3GwYu5lTH6
thbmKraNMlqW8ppeuYGgMP+gWGDbQfWgoqrO70/PPy98p7Tu2aJYKinLQo7TfwtSWen6S3VdNzUip3Mm
s7kyA1lZQFvzL+83Wf4VAAD//2MzG58ZAgAA
`,
	},

	"/tmpls/login.go.html": {
		local:   "tmpls/login.go.html",
		size:    388,
		modtime: 1475586403,
		compressed: `
H4sIAAAJbogA/2RQ0WrDMAx8br5C6L3kB5K8DwYbbD+Q2GorsGVjy2yh9N/ntNnmsbe70+nQ6XoFSycW
AlRWRwi3W/cczvAkXZ2R2E3oGtsS7Hp3HYZTSB6Mm3MeMZZEx01AmI1ykBF7G2oSC4InvQQ74uvL2ztO
3aGuMjmbSTdyGFhiUdA10ohKn4ogs6+4ZEobQohuNnQJzlJq5P/bsR7zEZL9TvjlfxJ+5EfCUlSD7BG5
LJ7rCW2x3dDgY0zs57TCTnMxhnLGKfNZgGXoH4N7276pW0n90tR89ysAAP//TNbcHIQBAAA=
`,
	},

	"/tmpls/stopped.go.html": {
		local:   "tmpls/stopped.go.html",
		size:    56,
		modtime: 1476223779,
		compressed: `
H4sIAAAJbogA/6quVkhJTcvMS1VQSspPqVRSqK3lUlCwyTC0C04tKkstUgguyS8oSE2x0QcKcQFVp+al
gNQAAgAA///xp9MCOAAAAA==
`,
	},

	"/": {
		isDir: true,
		local: "/",
	},

	"/public": {
		isDir: true,
		local: "/public",
	},

	"/public/css": {
		isDir: true,
		local: "/public/css",
	},

	"/public/js": {
		isDir: true,
		local: "/public/js",
	},

	"/tmpls": {
		isDir: true,
		local: "/tmpls",
	},
}
