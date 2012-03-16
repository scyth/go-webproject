package gwp_template

import (
	"html/template"
	"github.com/scyth/go-webproject/gwp/gwp_context"
)

// Load is API call which will return parsed template object, and will do this fast.
// It is also thread safe
func Load(ctx *gwp_context.Context, name string) (tpl *template.Template, err error) {
	if ctx.Templates[ctx.App.TemplatePath+name] != nil {
		return ctx.Templates[ctx.App.TemplatePath+name], nil
	}

	tpl, err = template.ParseFiles(ctx.App.TemplatePath + name)
	if err != nil {
		return nil, err
	}
	pt := &gwp_context.ParsedTemplate{ctx.App.TemplatePath + name, tpl}

	ctx.LiveTplMsg <- pt
	return tpl, nil
}
