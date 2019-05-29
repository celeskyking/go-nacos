package properties

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/sirupsen/logrus"
	"go-nacos/config"
	"go-nacos/config/listener"
	"go-nacos/config/types"
	"go-nacos/err"
	"go-nacos/pkg/pool"
	"os"
	"strconv"
	"strings"
)

func init() {
	config.RegisterConverter("properties", func(desc *types.FileDesc, content []byte) types.FileMirror {
		f, er := NewMapFile(desc,content)
		if er != nil {
			logrus.Errorf("load map file converter failed:%+v", er)
			os.Exit(1)
		}
		return f
	})
}


type Change struct {

	Key string

	EventType listener.EventType

	OldValue string

	NewValue string
}


func NewMapFile(desc *types.FileDesc, content []byte) (*MapFile, error) {
	f := &MapFile{Init:true}
	er := refresh(f, desc, content)
	if er != nil {
		return nil, er
	}
	f.Init = false
	f.valueListeners = make(map[string]listener.ValueListener)
	f.fileListeners = make([]listener.FileListener,0)
	return f, nil
}


func refresh(file *MapFile, desc *types.FileDesc, content []byte) error{
	m := md5.New()
	m.Write(content)
	md5Value := hex.EncodeToString(m.Sum(nil))
	params, er := toMap(content)
	if er != nil {
		return er
	}
	file.md5 = md5Value
	oldContent := file.content
	oldParams := file.params
	file.params = params
	file.content = content
	if !file.Init {
		diffFile(file, oldContent, content)
		diffParam(file,oldParams,params)
	}
	return nil
}



func diffParam(file *MapFile, oldValue, newValue map[string]string){
	if len(file.valueListeners) == 0 {
		return
	}
	changes := make([]*Change, 4)
	changes = append(changes, diffDeleted(oldValue,newValue)...)
	changes = append(changes, diffAdded(oldValue, newValue)...)
	for _, c := range changes {
		if v,ok := file.valueListeners[c.Key];ok {
			pool.Go(func(context context.Context) {
				v.OnChange(c.Key, c.OldValue,c.NewValue,file.Desc())
			})
		}
	}
}


func diffFile(file *MapFile, oldContent,newContent []byte) {
	if len(file.fileListeners) == 0 {
		return
	}
	for _, l := range file.fileListeners {
		pool.Go(func(i context.Context) {
			l.OnChange(oldContent,newContent,file.Desc())
		})
	}
}


func diffDeleted(oldValue, newValue map[string]string) []*Change{
	changes := make([]*Change, 0)
	if len(oldValue) == 0 {
		return changes
	}else if len(newValue) == 0 {
		for key, value := range oldValue{
			changes = append(changes, &Change{
				EventType:listener.Delete,
				Key:key,
				NewValue: value,
				OldValue: "",
			})
		}
	}else{
		for key, value := range oldValue {
			if v, ok := newValue[key]; ok {
				if v != value {
					changes = append(changes, &Change{
						EventType:listener.Update,
						Key:key,
						NewValue: v,
						OldValue:value,
					})
				}
			}else{
				changes = append(changes, &Change{
					EventType:listener.Delete,
					Key:key,
					NewValue: v,
					OldValue:"",
				})
			}
		}
	}
	return changes
}



func diffAdded(oldValue, newValue map[string]string) []*Change{
	changes := make([]*Change, 0)
	if len(oldValue) == 0 {
		for key, value := range newValue{
			changes = append(changes, &Change{
				EventType:listener.Add,
				Key:key,
				NewValue: value,
				OldValue: "",
			})
		}
		return changes
	}else if len(newValue) == 0 {
		return changes
	}else{
		for key, value := range newValue {
			if _, ok := oldValue[key]; !ok {
				changes = append(changes, &Change{
					EventType:listener.Add,
					Key:key,
					NewValue: value,
					OldValue: "",
				})
			}
		}
	}
	return changes
}







func toMap(content []byte) (map[string]string,error){
	m := make(map[string]string,8)
	if len(content) == 0 {
		return m, nil
	}
	reader := bufio.NewScanner(bytes.NewReader(content))
	for reader.Scan() {
		line := reader.Text()
		key,value, er := parseLine(line)
		if er !=nil {
			return nil, er
		}
		m[key] = value
	}
	return m, nil
}


func parseLine(line string) (string,string, error) {
	parts := strings.Split(line,"=")
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}else{
		return "", "", err.ErrNotPropertiesFile
	}
}

//对应properties文件
type MapFile struct{

	params map[string]string

	//根据namespace:env:appName 计算出来的hash值,md5，忽略碰撞的情况
	desc *types.FileDesc

	//文件的原值
	content []byte

	//文件变更的监听器
	fileListeners []listener.FileListener

	//值监听器
	valueListeners map[string]listener.ValueListener

	//文件内容的md5值
	md5 string

	//是否是初始化
	Init bool

}


func(m *MapFile) OnChanged(notifyC <- chan []byte) {
	for data := range notifyC {
		er :=  refresh(m,m.desc,data)
		if er != nil {
			logrus.Errorf("接受nacos 配置文件更新失败,error:%+v",er)
		}
	}
}


func(m *MapFile) GetContent() []byte {
	return m.content
}


func(m *MapFile) Desc() *types.FileDesc {
	return m.desc
}

func(m *MapFile) MD5() string {
	return m.md5
}


func(m *MapFile) Listen(f listener.FileListenerFunc) {
	m.fileListeners = append(m.fileListeners,f)
}


func(m *MapFile) ListenValue(key string, f listener.ValueListenerFunc) {
	m.valueListeners[key] = f
}


func(m *MapFile) Get(key string) (string,bool) {
	v, ok := m.params[key]
	return v, ok
}



func(m *MapFile) MustGet(key string) string {
	return m.params[key]
}


func(m *MapFile) GetBool(key string) (bool, error){
	v, ok := m.Get(key)
	if ok {
		return strconv.ParseBool(v)
	}
	return false, err.ErrkeyNotFound
}


func(m *MapFile) GetFloat32(key string) (float32,error) {
	v, ok := m.Get(key)
	if ok {
		f, er :=  strconv.ParseFloat(v, 32)
		return float32(f), er
	}
	return 0, err.ErrkeyNotFound
}


func(m *MapFile) MustGetFloat32(key string) float32 {
	v, ok := m.Get(key)
	if ok {
		f, er :=  strconv.ParseFloat(v, 32)
		if er != nil {
			panic(er)
		}
		return float32(f)
	}
	return 0
}


func(m *MapFile) MustGetFloat64(key string) float64 {
	v, ok := m.Get(key)
	if ok {
		f, er :=  strconv.ParseFloat(v, 32)
		if er != nil {
			panic(er)
		}
		return float64(f)
	}
	return 0
}


func(m *MapFile) GetFloat64(key string) (float64,error) {
	v, ok := m.Get(key)
	if ok {
		f, er :=  strconv.ParseFloat(v, 64)
		return f, er
	}
	return 0, err.ErrkeyNotFound
}


func(m *MapFile) MustGetBool(key string) bool {
	v, ok := m.Get(key)
	if ok {
		r, er := strconv.ParseBool(v)
		if er != nil {
			panic(er)
		}
		return r
	}
	return false
}

//GetInt 返回int类型的值,如果值不存在则抛出异常
func(m *MapFile) GetInt(key string) (int, error) {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 32)
		return int(i), er
	}
	return 0, err.ErrkeyNotFound
}

func(m *MapFile) MustGetInt(key string) int {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 32)
		if er != nil {
			panic(er)
		}
		return int(i)
	}
	return 0
}

func(m *MapFile) GetInt32(key string) (int32,error) {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 32)
		return int32(i), er
	}
	return 0, err.ErrkeyNotFound
}



func(m *MapFile) MustGetInt32(key string) int32 {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 32)
		if er != nil {
			panic(er)
		}
		return int32(i)
	}
	return 0
}


func(m *MapFile) GetInt64(key string)( int64,error) {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 64)
		return i, er
	}
	return 0, err.ErrkeyNotFound
}

func(m *MapFile) MustGetInt64(key string) int64 {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 64)
		if er != nil {
			panic(er)
		}
		return i
	}
	return 0
}

func(m *MapFile) GetUint32(key string) (uint32,error) {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 32)
		return uint32(i), er
	}
	return 0, err.ErrkeyNotFound
}

func(m *MapFile) MustGetUint32(key string) uint32 {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 32)
		if er != nil {
			panic(er)
		}
		return uint32(i)
	}
	return 0
}


func(m *MapFile) GetUint64(key string) (uint64,error) {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 64)
		return uint64(i), er
	}
	return 0, err.ErrkeyNotFound
}

func(m *MapFile) MustGetUint64(key string) uint64 {
	v, ok := m.Get(key)
	if ok {
		i, er := strconv.ParseInt(v, 10, 64)
		if er != nil {
			panic(er)
		}
		return uint64(i)
	}
	return 0
}