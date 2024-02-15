package extension

import (
	"bytes"
	"context"
	"fmt"
	"gorm.io/gen"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type TemplateConfig struct {
	FuncM template.FuncMap       // 自定义模版函数
	Data  map[string]interface{} // 自定义模版数据
	// TODO more ...
}

// 加载模版
func loadTemplate(basePath string) ([]string, error) {
	var result []string
	/*
	 * 加载路径条目
	 */
	dir, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, value := range dir {
		n := value.Name()
		if !value.IsDir() {
			/*
			 * 只获取tpl后缀的文件
			 */
			if !strings.HasSuffix(n, ".tpl") {
				continue
			}

			/*
			 * 读取模版内容
			 */
			filename := fmt.Sprintf("%s%c%s", basePath, os.PathSeparator, n)
			data, erro := os.ReadFile(filename)
			if erro != nil {
				return nil, erro
			}
			result = append(result, string(data))
		} else {
			/*
			 * 进入下级目录
			 */
			next := fmt.Sprintf("%s%c%s", basePath, os.PathSeparator, n)
			templates, erro := loadTemplate(next)
			if erro != nil {
				return nil, erro
			}
			result = append(result, templates...)
		}
	}
	return result, nil
}

func BuildTemplateExtension(basePath string, config TemplateConfig) gen.Extension {
	if config.Data == nil {
		config.Data = make(map[string]interface{})
	}

	return func(
		g *gen.Generator,
		render func(tmpl string, w io.Writer, data interface{}, funcM template.FuncMap) error,
		output func(fileName string, content []byte) error,
	) (err error) {
		g.Logger().Info(context.Background(), "BuildTemplateExtension ...")
		log.Println("BuildTemplateExtension ...")
		templates, err := loadTemplate(basePath)
		if err != nil {
			return err
		}

		const E = ""

		for _, tpl := range templates {
			for _, value := range g.Data {
				var (
					buf      bytes.Buffer // 模版输出缓存
					filename string       // 输出文件
					ignore   bool         // 跳过生成文件
				)

				/*
				 * 内置模版函数
				 */
				funcM := template.FuncMap{
					"output": func(value string) string {
						filename = value
						return E
					},
					"ignore": func() string {
						ignore = true
						return E
					},
				}
				if config.FuncM != nil {
					for k, v := range config.FuncM {
						funcM[k] = v
					}
				}

				/*
				 * 渲染模板
				 */
				err = render(tpl, &buf, map[string]interface{}{
					"Gen":  value,
					"Data": config.Data,
				}, funcM)
				if err != nil {
					return
				}

				// 跳过本次生成的模板
				if ignore {
					continue
				}

				/*
				 * 保存文件
				 */
				if filename != "" {
					g.Logger().Info(context.Background(), "Generate: ", filename)
					log.Println("Generate: ", filename)
					dir := filepath.Dir(filename)
					if err = os.MkdirAll(dir, os.ModePerm); err != nil {
						return fmt.Errorf("make dir outpath(%s) fail: %s", g.OutPath, err)
					}
					err = output(filename, buf.Bytes())
				} else {
					return fmt.Errorf("we need call `output` mothod to save filename")
				}

			}
		}

		return nil
	}
}
