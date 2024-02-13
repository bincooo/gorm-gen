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

func BuildTemplateExtension(basePath string) gen.Extension {
	var loadTemplate func(basePath string) ([]string, error)
	//
	//

	// 加载模版
	loadTemplate = func(basePath string) ([]string, error) {
		dir, err := os.ReadDir(basePath)
		if err != nil {
			return nil, err
		}

		var result []string
		for _, f := range dir {
			n := f.Name()
			if !f.IsDir() {
				if !strings.HasSuffix(n, ".tpl") {
					continue
				}
				filename := fmt.Sprintf("%s%c%s", basePath, os.PathSeparator, n)
				data, err := os.ReadFile(filename)
				if err != nil {
					return nil, err
				}
				result = append(result, string(data))
			} else {
				next := fmt.Sprintf("%s%c%s", basePath, os.PathSeparator, n)
				templates, err := loadTemplate(next)
				if err != nil {
					return nil, err
				}
				result = append(result, templates...)
			}
		}
		return result, nil
	}

	//
	//
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
					buf      bytes.Buffer //
					filename string       // 输出文件
				)

				funcM := template.FuncMap{
					"output": func(value string) string {
						filename = value
						return E
					},
				}

				err = render(tpl, &buf, value, funcM)
				if err != nil {
					return
				}

				if filename != "" {
					g.Logger().Info(context.Background(), "Generate: ", filename)
					log.Println("Generate: ", filename)
					dir := filepath.Dir(filename)
					if err = os.MkdirAll(dir, os.ModePerm); err != nil {
						return fmt.Errorf("make dir outpath(%s) fail: %s", g.OutPath, err)
					}
					err = output(filename, buf.Bytes())
				}
			}
		}

		return nil
	}
}
