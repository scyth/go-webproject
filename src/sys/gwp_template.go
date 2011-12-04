package gwp_template

import (
	"gwp/gwp_context"
	"html/template"
)

// Load is API call which will return parsed template object, and will do this fast.
// It is also thread safe
func Load(ctx *gwp_context.Context, name string) (tpl *template.Template, err error) {
	if ctx.App.Templates[ctx.App.TemplatePath+name] != nil {
		return ctx.App.Templates[ctx.App.TemplatePath+name], nil
	}

	tpl, err = template.ParseFiles(ctx.App.TemplatePath + name)
	if err != nil {
		return nil, err
	}
	pt := &gwp_context.ParsedTemplate{ctx.App.TemplatePath + name, tpl}

	ctx.LiveTplMsg <- pt
	return tpl, nil
}
