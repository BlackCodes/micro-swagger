package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/logger"
	"github.com/tidwall/pretty"
)

var Storage string
var lock sync.RWMutex

const (
	BaseFile = "base.json"
)

type Swagger struct {
	prePath  string
}

func NewSwagger() *Swagger {
	return &Swagger{}
}
func (s *Swagger) Get(c *gin.Context) {

	fileName, _ := c.GetQuery("file")
	project := c.Param("project")

	if len(fileName) == 0 {
		fileName = BaseFile
	}

	path := s.getPath(project, fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Infof("Not found Swagger:%v,path:%v", fileName, path)
		c.Writer.WriteString(fmt.Sprintf("Not fond swagger file:%s", path))
		return
	}
	f, err := os.Open(path)
	if err != nil {
		logger.Errorf("open the file err %v", err)
		return
	}
	b, _ := ioutil.ReadAll(f)
	b = pretty.Pretty(b)
	c.Writer.Write(b)
	c.Writer.Flush()
}

func (s *Swagger) Push(c *gin.Context) {

	var req struct {
		FileName string `json:"fileName"`
		Content  string `json:"content"`
		Project  string `json:"project"`
	}

	if err := c.BindJSON(&req); err != nil {
		logger.Errorf("bind json err:%v", err)
		return
	}
	logger.Infof("the push file ProjectName:%s,FileName:%v", req.Project, req.FileName)
	dir := s.getPath("", "")
	if len(req.Project) == 0 {
		c.AbortWithError(504, fmt.Errorf("project name empty"))
		return
	}
	dir = s.getPath(req.Project, "")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0766)
	}
	swagger, err := s.parseSwagger([]byte(req.Content))
	if err != nil {
		return
	}
	s.detectPrePath(swagger)
	swagger = s.setUnion(swagger)
	newStoreName := s.getNewFileName(swagger.Paths, req.FileName)
	lock.Lock()
	defer lock.Unlock()
	mergeSwg, err := s.mergeMsg(req.Project, newStoreName, swagger)
	if err != nil {
		logger.Errorf("mergeMsg swagger json err:%v", err)
		return
	}
	path := s.getPath(req.Project, BaseFile)
	if err := s.WriteTo(path, mergeSwg); err != nil {
		logger.Errorf("write to file %v", err)
		return
	}
	newSwagerPath := s.getPath(req.Project, newStoreName)
	if err := s.WriteTo(newSwagerPath, swagger); err != nil {
		logger.Errorf("write to file %v", err)
		return
	}
	logger.Info("success merge")
}

func (s *Swagger) parseSwagger(c []byte) (*openapiSwaggerObject, error) {
	var swagger *openapiSwaggerObject
	if err := json.Unmarshal(c, &swagger); err != nil {
		logger.Errorf("json unmarshal err %v", err)
		return nil, err
	}
	return swagger, nil
}

func (s *Swagger) loadSwg(filePath string) (*openapiSwaggerObject, error) {
	f, err := os.Open(filePath)
	if err != nil {
		logger.Warnf("open base file err:%v", err)
		return nil, nil
	}
	defer f.Close()
	b, _ := ioutil.ReadAll(f)
	if len(b) == 0 {
		logger.Info("the base file is empty")
		return nil, nil
	}
	objSwg, err := s.parseSwagger(b)
	if err != nil {
		return nil, err
	}
	return objSwg, nil

}

func (s *Swagger) setUnion(swg *openapiSwaggerObject) *openapiSwaggerObject {
	if len(s.prePath) == 0 {
		return swg
	}
	modify := make(map[string]string )
	for _,path := range swg.Paths{
		for _,item := range path.Post.Parameters{
			if item.Schema != nil &&  !strings.Contains(item.Schema.Ref,s.prePath) {
				newKey := fmt.Sprintf("%s_%s",item.Schema.Ref,s.prePath)
				lastKey := item.Schema.Ref[strings.LastIndex(item.Schema.Ref,"/")+1:]
				modify[lastKey] = fmt.Sprintf("%s_%s",lastKey,s.prePath)
				item.Schema.Ref = newKey
			}
		}
		for key, item := range path.Post.Responses{
			if !strings.Contains(item.Schema.Ref,s.prePath) {
				newKey := fmt.Sprintf("%s_%s",item.Schema.Ref,s.prePath)
				lastKey := item.Schema.Ref[strings.LastIndex(item.Schema.Ref,"/")+1:]
				modify[lastKey] = fmt.Sprintf("%s_%s",lastKey,s.prePath)
				item.Schema.Ref = newKey
				path.Post.Responses[key] = item
			}
		}
	}
	for key ,item := range swg.Definitions{
		if len(item.Ref) > 0 {
			refKey := item.Ref[strings.LastIndex(item.Ref,"/"):]
			if s,ok := modify[refKey];ok {
				item.Ref = s
			}
		}
		if s,ok := modify[key];ok {
			swg.Definitions[s] = item
			delete(swg.Definitions,key)
		}
	}
	return swg
}

func (s *Swagger) detectPrePath(swg *openapiSwaggerObject)   {
	k := ""
	for key,_ := range swg.Paths{
		arr := strings.Split(key,"/" )
		if len(k) == 0 && len(arr)!=0 {
			k = arr[1]
			break
		}
	}
	s.prePath = k
}
func (s *Swagger) mergeMsg(project string, newPathName string, swg *openapiSwaggerObject) (*openapiSwaggerObject, error) {
	basePath := s.getPath(project, BaseFile)
	baseSwg, err := s.loadSwg(basePath)
	if err != nil {
		return nil, err
	}
	if baseSwg == nil {
		baseSwg = swg
	}
	delPaths, delDefs, err := s.CheckDelete(s.getPath(project, newPathName), swg)
	if err != nil {
		return nil, err
	}
	definitions := s.DefinitionMerge(delDefs, baseSwg.Definitions, swg.Definitions)
	paths := s.PathsMerge(delPaths, baseSwg.Paths, swg.Paths)
	baseSwg.Paths = *paths
	baseSwg.Definitions = *definitions
	return baseSwg, nil
}

func (s *Swagger) DefinitionMerge(delDefs map[string]struct{}, base, new openapiDefinitionsObject) *openapiDefinitionsObject {
	if base == nil {
		base = make(openapiDefinitionsObject)
	}
	newBase := make(openapiDefinitionsObject)
	for def, item := range base {
		if _, ok := delDefs[def]; !ok {
			newBase[def] = item
		} else {
			logger.Infof("current merge will del definition:%s", def)
		}
	}
	for nk, nv := range new {
		isFound := false
		for k, _ := range newBase {
			if nk == k {
				isFound = true
				newBase[k] = nv
				break
			}
		}
		if !isFound {
			newBase[nk] = nv
		}
	}
	return &newBase
}

func (s *Swagger) PathsMerge(delPath map[string]struct{}, base, new openapiPathsObject) *openapiPathsObject {
	if base == nil {
		base = make(openapiPathsObject)
	}
	newBase := make(openapiPathsObject)
	for def, item := range base {
		if _, ok := delPath[def]; !ok {
			newBase[def] = item
		} else {
			logger.Infof("current merge will del path:%s", def)
		}
	}
	for nk, nv := range new {
		isFound := false
		for k, _ := range newBase {
			if nk == k {
				isFound = true
				newBase[k] = nv
				break
			}
		}
		if !isFound {
			newBase[nk] = nv
		}
	}
	return &newBase
}

func (s *Swagger) WriteTo(path string, swg *openapiSwaggerObject) error {
	logger.Info("write to path", path)
	b, err := json.Marshal(swg)
	if err != nil {
		return err
	}
	var _swg *openapiSwaggerObject
	if err := json.Unmarshal(b, &_swg); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	_,err = f.Write(b)
	if err !=nil {
		f.Sync()
	}
	return err
}

func (s *Swagger) getPath(projectName, fileName string) string {
	p := strings.TrimRight(Storage, "/")
	if len(projectName) > 0 {
		p = fmt.Sprintf("%s/%s", p, projectName)
	}
	if len(fileName) > 0 {
		p = fmt.Sprintf("%s/%s", p, fileName)
	}
	return p
}

func (s *Swagger) getNewFileName(paths openapiPathsObject, fileName string) string {
	p := ""
	for _p, _ := range paths {
		arrP := strings.Split(_p, "/")
		p = _p
		if len(arrP) > 1 {
			arrNew := arrP[:len(arrP)-1]
			p = strings.TrimLeft(strings.Join(arrNew, "_"), "_")
		}
		break
	}
	if len(p) == 0 {
		return filepath.Base(fileName)
	}
	return fmt.Sprintf("%s_%s", p, filepath.Base(fileName))
}

func (s *Swagger) CheckDelete(file string, swg *openapiSwaggerObject) (delPath, delDef map[string]struct{}, err error) {
	old, err := s.loadSwg(file)
	if err != nil {
		return nil, nil, err
	}
	if old == nil {
		// empty
		return nil, nil, nil
	}
	delPath = make(map[string]struct{})
	for oldKey, _ := range old.Paths {
		isFond := false
		for newKey, _ := range swg.Paths {
			if newKey == oldKey {
				isFond = true
				break
			}
		}
		if !isFond {
			delPath[oldKey] = struct{}{}
		}
	}

	delDef = make(map[string]struct{})
	for oldKey, _ := range old.Definitions {
		isFound := false
		for newKey, _ := range swg.Definitions {
			if oldKey == newKey {
				isFound = true
				break
			}
		}
		if !isFound {
			delDef[oldKey] = struct{}{}
		}
	}
	return delPath, delDef, nil

}
