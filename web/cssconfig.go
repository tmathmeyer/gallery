package web

import (
	"database/sql"
	"html/template"
	"net/http"
	"log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type CssConfiguration struct {
	Background			string `yaml:"background"`
	HeaderBackground	string `yaml:"header_background"`
	TextColorPress		string `yaml:"text_color_press"`
	ShadowColorPress	string `yaml:"shadow_color_press"`
	BreadcrumbColor1	string `yaml:"breadcrumb_color_1"`
	BreadcrumbText1		string `yaml:"breadcrumb_text_1"`
	BreadcrumbColor2	string `yaml:"breadcrumb_color_2"`
	BreadcrumbText2		string `yaml:"breadcrumb_text_2"`
	BreadcrumbColor3	string `yaml:"breadcrumb_color_3"`
	BreadcrumbText3		string `yaml:"breadcrumb_text_3"`
	BreadcrumbColor4	string `yaml:"breadcrumb_color_4"`
	BreadcrumbText4		string `yaml:"breadcrumb_text_4"`
	BreadcrumbColor5	string `yaml:"breadcrumb_color_5"`
	BreadcrumbText5		string `yaml:"breadcrumb_text_5"`
}

func (c *CssConfiguration) getColors() *CssConfiguration {
    yamlFile, err := ioutil.ReadFile("colors.yaml")
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }

    return c
}

func CssConfigureHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("templates/common.css")
		var conf CssConfiguration
		conf.getColors()
		t.Execute(w, conf)
	})
}