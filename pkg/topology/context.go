package topology

import (
	"github.com/flanksource/canary-checker/api/context"
	v1 "github.com/flanksource/canary-checker/api/v1"
	"github.com/flanksource/canary-checker/pkg"
	"github.com/flanksource/kommons"
	"github.com/flanksource/kommons/ktemplate"
	"github.com/pkg/errors"
)

type ComponentContext struct {
	*context.KubernetesContext
	SystemTemplate v1.SystemTemplate
	ComponentAPI   v1.Component
	// Components keep track of the components that properties can apply to,
	// properties can return a map of component names to properties to facilitate
	// queries that are more efficient to perform for all components rather than a component at a time
	Components *pkg.Components
	// Properties can either be looked up on an individual component, or act as a summary across all components
	CurrentComponent *pkg.Component
	templater        *ktemplate.StructTemplater
}

func (c *ComponentContext) GetTemplater() ktemplate.StructTemplater {
	if c.templater != nil {
		return *c.templater
	}
	c.templater = &ktemplate.StructTemplater{
		// RequiredTag: "template",
		DelimSets: []ktemplate.Delims{
			{
				Left:  "${",
				Right: "}",
			},
		},
		Values: map[string]interface{}{
			"component":  c.CurrentComponent,
			"properties": c.CurrentComponent.Properties.AsMap(),
		},
	}
	return *c.templater
}

func (c *ComponentContext) SetCurrentComponent(component *pkg.Component) {
	c.CurrentComponent = component
	if c.templater != nil {
		c.templater.Values = map[string]interface{}{
			"component":  c.CurrentComponent,
			"properties": c.CurrentComponent.Properties.AsMap(),
		}
	}
}

func (c *ComponentContext) TemplateProperty(property *v1.Property) error {
	templater := c.GetTemplater()
	if err := templater.Walk(property); err != nil {
		return errors.Wrapf(err, "failed to template property %s", property.Name)
	}
	return nil
}

func (c *ComponentContext) TemplateStruct(data interface{}) error {
	templater := c.GetTemplater()
	if err := templater.Walk(data); err != nil {
		return errors.Wrapf(err, "failed to template struct %s", data)
	}
	return nil
}

func (c *ComponentContext) TemplateConfig(config *v1.Config) error {
	templater := c.GetTemplater()
	if err := templater.Walk(config); err != nil {
		return errors.Wrapf(err, "failed to template config %s", *config)
	}
	//FIXME struct templater does not support maps
	var labels = make(map[string]string)
	for k, v := range config.Labels {
		labels[k], _ = templater.Template(v)
	}
	(*config).Labels = labels
	return nil
}

func (c *ComponentContext) TemplateComponent(component *v1.ComponentSpec) error {
	templater := c.GetTemplater()
	if err := templater.Walk(component); err != nil {
		return errors.Wrapf(err, "failed to template component %s", *component)
	}
	return nil
}

func (c *ComponentContext) Clone() *ComponentContext {
	return &ComponentContext{
		KubernetesContext: c.KubernetesContext.Clone(),
		SystemTemplate:    c.SystemTemplate,
		ComponentAPI:      c.ComponentAPI,
		Components:        c.Components,
	}
}
func (c *ComponentContext) WithComponents(components *pkg.Components, current *pkg.Component) *ComponentContext {
	cloned := c.Clone()
	cloned.Components = components
	cloned.CurrentComponent = current
	return cloned
}

func NewComponentContext(client *kommons.Client, system v1.SystemTemplate) *ComponentContext {
	return &ComponentContext{
		KubernetesContext: context.NewKubernetesContext(client, system.Namespace),
		SystemTemplate:    system,
	}
}
