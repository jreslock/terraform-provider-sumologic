package sumologic

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceSumologicField() *schema.Resource {
	return &schema.Resource{
		Create: resourceSumologicFieldCreate,
		Read:   resourceSumologicFieldRead,
		Update: resourceSumologicFieldUpdate,
		Delete: resourceSumologicFieldDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSumologicFieldImport,
		},

		Schema: map[string]*schema.Schema{

			"field_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"field_id": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: false,
			},

			"data_type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceSumologicFieldRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)

	id := d.Get("field_id").(string)
	name := d.Get("field_name").(string)
	if id == "" {
		newId, err := c.FindFieldId(name)
		if err != nil {
			return err
		}
		id = newId
	}

	field, err := c.GetField(id)
	if err != nil {
		return err
	}

	if field == nil {
		fmt.Printf("[WARN] Field not found, removing from state: %v - %v\n", id, err)
		d.SetId("")
		return nil
	}

	d.Set("field_name", field.FieldName)
	d.Set("field_id", field.FieldId)
	d.Set("data_type", field.DataType)
	d.Set("state", field.State)

	return nil
}

func resourceSumologicFieldDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)

	id := d.Get("field_id").(string)
	name := d.Get("field_name").(string)
	if id == "" {
		newId, err := c.FindFieldId(name)
		if err != nil {
			return err
		}
		id = newId
	}

	return c.DeleteField(id)
}

func resourceSumologicFieldCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)

	if d.Id() == "" {
		field := resourceToField(d)
		id, err := c.CreateField(field)
		if err != nil {
			return err
		}

		d.SetId(id)
	}

	return resourceSumologicFieldRead(d, meta)
}

func resourceSumologicFieldImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Get("field_id").(string) == "" {
		d.Set("field_id", d.Id())
	}

	return []*schema.ResourceData{d}, nil
}

func resourceSumologicFieldUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)

	id := d.Get("field_id").(string)
	name := d.Get("field_name").(string)
	tpe := d.Get("data_type").(string)
	status := d.Get("state").(string)
	if id == "" {
		newId, err := c.FindFieldId(name)
		if err != nil {
			return err
		}
		id = newId
	}
	f, err := c.GetField(id)

	if err != nil {
		return err
	}

	if f.FieldName != name && f.DataType != tpe {
		return errors.New("Only state field is updatable")
	}

	if status == "Enabled" {
		return c.EnableField(id)
	} else if status == "Disabled" {
		return c.DisableField(id)
	} else {
		return errors.New("Invalid value of state field. Only Enabled or Disabled values are accepted")
	}
}

func resourceToField(d *schema.ResourceData) Field {
	return Field{
		DataType:  d.Get("data_type").(string),
		FieldId:   d.Get("field_id").(string),
		State:     d.Get("state").(string),
		FieldName: d.Get("field_name").(string),
	}
}
