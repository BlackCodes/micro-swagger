package handler

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPush(t *testing.T) {
	path := "/Users/jeff.tian/Desktop/go/src/elysium/doc/web/service.swagger.json"
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("open file err:", err)
	}
	b, _ := ioutil.ReadAll(f)
	//fmt.Println("data",string(b))
	defer f.Close()
	swg := NewSwagger()
	newSwg, err := swg.parseSwagger(b)
	t.Run("Parse", func(t *testing.T) {
		for k,v := range newSwg.Definitions{
			fmt.Println("Key:",k)
			fmt.Println("v:",v.Properties)
		}
	})
	if err != nil {
		t.Errorf("parse swagger err:%v", err)
		return
	}

	t.Run("加载Base", func(t *testing.T) {
		_, err := swg.loadSwg()
		if err != nil {
			t.Errorf("load base  swagger err:%v", err)
			return
		}

	})
	t.Run("整体流程", func(t *testing.T) {
		t.Run("合并swagger", func(t *testing.T) {
			fileSwg, err := swg.mergeMsg(newSwg)
			if err != nil {
				t.Errorf("mergeMsg to base swagger err:%v", err)
				return
			}

			t.Run("生成文件", func(t *testing.T) {
				path := fmt.Sprintf("%s/%s",Storage,BaseFile)
				if err := swg.WriteTo(path,fileSwg); err != nil {
					t.Errorf("write to file err:%v", err)
					return
				}
				t.Log("Perfect...")
			})
		})

	})

}

func TestSwagger_Get(t *testing.T) {
	r := &httptest.ResponseRecorder{}
	req := httptest.NewRequest("GET","http://127.0.0.1:9099?fileName=tt.json",nil )
	c,_ := gin.CreateTestContext(r)

	c.Request = req
	NewSwagger().Get(c)
	t.Log(r.Body.String())
}