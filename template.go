package lemon

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Template struct {
	FuncMap template.FuncMap
	// beego template caching map and supported template file extensions.
	Templates    map[string]*template.Template
	LeftBraces   string
	RightBraces  string
	templatefile *TemplateFile
}

type TemplateFile struct {
	root  string
	files map[string][]string
}

//var template *Template

func TemplateInit(leftBraces, rightBraces string) *Template {
	Templates := make(map[string]*template.Template)
	FuncMap := make(template.FuncMap)
	//TemplateExt = append(BeeTemplateExt, "tpl", "html")
	FuncMap["dateformat"] = DateFormat
	FuncMap["date"] = Date
	FuncMap["compare"] = Compare
	FuncMap["substr"] = Substr
	FuncMap["html2str"] = Html2str
	FuncMap["str2html"] = Str2html
	FuncMap["htmlquote"] = Htmlquote
	FuncMap["htmlunquote"] = Htmlunquote
	return &Template{FuncMap: FuncMap, Templates: Templates, LeftBraces: leftBraces, RightBraces: rightBraces}

}

// AddFuncMap let user to register a func in the template.
func (tpl *Template) AddFuncMap(key string, funname interface{}) error {
	tpl.FuncMap[key] = funname
	return nil
}

func (tpl *TemplateFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}

	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	//	if !HasTemplateExt(paths) {
	//		return nil
	//	}
	replace := strings.NewReplacer("\\", "/")
	a := []byte(paths)
	a = a[len([]byte(tpl.root)):]
	file := strings.TrimLeft(replace.Replace(string(a)), "/")
	subdir := filepath.Dir(file)
	if _, ok := tpl.files[subdir]; ok {
		tpl.files[subdir] = append(tpl.files[subdir], file)
	} else {
		m := make([]string, 1)
		m[0] = file
		tpl.files[subdir] = m
	}

	return nil
}

func (tpl *Template) BuildTemplate(templatePath string) error {
	//	workPath, _ := os.Getwd()
	//	AbsWorkPath, _ := filepath.Abs(workPath)
	//	templatePath := filepath.Join(AbsWorkPath, dir)
	if _, err := os.Stat(templatePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return errors.New("dir open err")
		}
	}
	tpl.templatefile = &TemplateFile{
		root:  templatePath,
		files: make(map[string][]string),
	}
	err := filepath.Walk(templatePath, func(path string, f os.FileInfo, err error) error {
		return tpl.templatefile.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}
	for _, v := range tpl.templatefile.files {
		for _, file := range v {
			t, err := tpl.getTemplate(tpl.templatefile.root, file, v...)
			if err != nil {
				//Trace("parse template err:", file, err)
			} else {
				tpl.Templates[file] = t
			}
		}
	}
	return nil
}

func (tpl *Template) getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileabspath string
	if filepath.HasPrefix(file, "../") {
		fileabspath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileabspath = filepath.Join(root, file)
	}
	if e := tpl.FileExists(fileabspath); !e {
		panic("can't find template file:" + file)
	}
	data, err := ioutil.ReadFile(fileabspath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile(tpl.LeftBraces + "[ ]*template[ ]+\"([^\"]+)\"")
	allsub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allsub {
		if len(m) == 2 {
			tlook := t.Lookup(m[1])
			if tlook != nil {
				continue
			}
			//			if !HasTemplateExt(m[1]) {
			//				continue
			//			}
			t, _, err = tpl.getTplDeep(root, m[1], file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allsub, nil
}

func (tpl *Template) getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file).Delims(tpl.LeftBraces, tpl.RightBraces).Funcs(tpl.FuncMap)
	var submods [][]string
	t, submods, err = tpl.getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = tpl._getTemplate(t, root, submods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func (tpl *Template) _getTemplate(t0 *template.Template, root string, submods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range submods {
		if len(m) == 2 {
			templ := t.Lookup(m[1])
			if templ != nil {
				continue
			}
			//first check filename
			for _, otherfile := range others {
				if otherfile == m[1] {
					var submods1 [][]string
					t, submods1, err = tpl.getTplDeep(root, otherfile, "", t)
					if err != nil {
						//Trace("template parse file err:", err)
					} else if submods1 != nil && len(submods1) > 0 {
						t, err = tpl._getTemplate(t, root, submods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherfile := range others {
				fileabspath := filepath.Join(root, otherfile)
				data, err := ioutil.ReadFile(fileabspath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile(tpl.LeftBraces + "[ ]*define[ ]+\"([^\"]+)\"")
				allsub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allsub {
					if len(sub) == 2 && sub[1] == m[1] {
						var submods1 [][]string
						t, submods1, err = tpl.getTplDeep(root, otherfile, "", t)
						if err != nil {
							//Trace("template parse file err:", err)
						} else if submods1 != nil && len(submods1) > 0 {
							t, err = tpl._getTemplate(t, root, submods1, others...)
						}
						break
					}
				}
			}
		}

	}
	return
}

func (tpl *Template) FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func AddFuncMap(funcMaps map[string]interface{}) map[string]interface{} {
	FuncMap := make(template.FuncMap, 0)
	for key, function := range funcMaps {
		FuncMap[key] = function
	}
	return FuncMap
}
