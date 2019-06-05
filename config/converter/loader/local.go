package loader

import (
	"github.com/celeskyking/go-nacos/config"
	"github.com/celeskyking/go-nacos/err"
	"github.com/celeskyking/go-nacos/pkg/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
)

type LocalLoader struct {
	//快照的目录
	SnapshotDir string
}

func NewLocalLoader(snapshotDir string) Loader {
	if snapshotDir == "" {
		panic(errors.New("快照目录不能够为空"))
	}
	return &LocalLoader{
		SnapshotDir: snapshotDir,
	}
}

//加载
func (ll *LocalLoader) Load(desc *config.FileDesc) ([]byte, error) {
	parts := []string{ll.SnapshotDir, desc.Namespace, desc.AppName, desc.Env, desc.Name}
	p := filepath.Join(parts...)
	if r, er := util.PathExists(p); er != nil {
		return nil, er
	} else {
		if r {
			data, er := ioutil.ReadFile(p)
			if er != nil {
				return nil, er
			}
			return data, nil
		} else {
			logrus.Errorf("file not found:%s", p)
			return nil, err.ErrFileNotFound
		}
	}
}

//向文件中写入内容
func (ll *LocalLoader) Write(desc *config.FileDesc, content []byte) error {
	parts := []string{ll.SnapshotDir, desc.Namespace, desc.AppName, desc.Env}
	p := filepath.Join(parts...)
	if e, er := util.PathExists(p); er != nil {
		return errors.Wrap(er, "snapshot dir exist")
	} else if !e {
		er := os.MkdirAll(p, 0755)
		if er != nil {
			return errors.Wrap(er, "create snapshot dir")
		}
	}
	logrus.Infof("flush config file to %s", p)
	return ioutil.WriteFile(filepath.Join(p, desc.Name), content, 0666)
}
