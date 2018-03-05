package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsSsmParameter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsmParameterPut,
		Read:   resourceAwsSsmParameterRead,
		Update: resourceAwsSsmParameterPut,
		Delete: resourceAwsSsmParameterDelete,
		Exists: resourceAwsSmmParameterExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateSsmParameterType,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"key_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"overwrite": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"allowed_pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSmmParameterExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	ssmconn := meta.(*AWSClient).ssmconn

	resp, err := ssmconn.GetParameters(&ssm.GetParametersInput{
		Names:          []*string{aws.String(d.Id())},
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		return false, err
	}
	return len(resp.InvalidParameters) == 0, nil
}

func resourceAwsSsmParameterRead(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[DEBUG] Reading SSM Parameter: %s", d.Id())

	resp, err := ssmconn.GetParameters(&ssm.GetParametersInput{
		Names:          []*string{aws.String(d.Id())},
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return errwrap.Wrapf("[ERROR] Error getting SSM parameter: {{err}}", err)
	}
	if len(resp.Parameters) == 0 {
		log.Printf("[WARN] SSM Param %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	param := resp.Parameters[0]
	d.Set("name", param.Name)
	d.Set("type", param.Type)
	d.Set("value", param.Value)

	describeParamsInput := &ssm.DescribeParametersInput{
		Filters: []*ssm.ParametersFilter{
			&ssm.ParametersFilter{
				Key:    aws.String("Name"),
				Values: []*string{aws.String(d.Get("name").(string))},
			},
		},
	}
	detailedParameters := []*ssm.ParameterMetadata{}
	err = ssmconn.DescribeParametersPages(describeParamsInput,
		func(page *ssm.DescribeParametersOutput, lastPage bool) bool {
			detailedParameters = append(detailedParameters, page.Parameters...)
			return !lastPage
		})
	if err != nil {
		return errwrap.Wrapf("[ERROR] Error describing SSM parameter: {{err}}", err)
	}
	if len(detailedParameters) == 0 {
		log.Printf("[WARN] SSM Param %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	detail := detailedParameters[0]
	if detail.Description != nil {
		// Trailing spaces are not considered as a difference
		*detail.Description = strings.TrimSpace(*detail.Description)
	}

	d.Set("key_id", detail.KeyId)
	d.Set("description", detail.Description)
	d.Set("allowed_pattern", detail.AllowedPattern)

	if tagList, err := ssmconn.ListTagsForResource(&ssm.ListTagsForResourceInput{
		ResourceId:   aws.String(d.Get("name").(string)),
		ResourceType: aws.String("Parameter"),
	}); err != nil {
		return fmt.Errorf("Failed to get SSM parameter tags for %s: %s", d.Get("name"), err)
	} else {
		d.Set("tags", tagsToMapSSM(tagList.TagList))
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "ssm",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("parameter/%s", strings.TrimPrefix(d.Id(), "/")),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceAwsSsmParameterDelete(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Deleting SSM Parameter: %s", d.Id())

	_, err := ssmconn.DeleteParameter(&ssm.DeleteParameterInput{
		Name: aws.String(d.Get("name").(string)),
	})
	if err != nil {
		return err
	}
	d.SetId("")

	return nil
}

func resourceAwsSsmParameterPut(d *schema.ResourceData, meta interface{}) error {
	ssmconn := meta.(*AWSClient).ssmconn

	log.Printf("[INFO] Creating SSM Parameter: %s", d.Get("name").(string))

	// Overwrite is set to true if the ressource already exists in the state or
	// if it has been explicitly specified
	paramInput := &ssm.PutParameterInput{
		Name:           aws.String(d.Get("name").(string)),
		Type:           aws.String(d.Get("type").(string)),
		Value:          aws.String(d.Get("value").(string)),
		Overwrite:      aws.Bool(!d.IsNewResource() || d.Get("overwrite").(bool)),
		AllowedPattern: aws.String(d.Get("allowed_pattern").(string)),
	}

	if description, ok := d.GetOk("description"); ok {
		paramInput.SetDescription(description.(string))
	} else if d.HasChange("description") {
		paramInput.SetDescription("")
	}

	if keyID, ok := d.GetOk("key_id"); ok {
		log.Printf("[DEBUG] Setting key_id for SSM Parameter %v: %s", d.Get("name"), keyID)
		paramInput.SetKeyId(keyID.(string))
	}

	log.Printf("[DEBUG] Waiting for SSM Parameter %v to be updated", d.Get("name"))
	if _, err := ssmconn.PutParameter(paramInput); err != nil {
		return errwrap.Wrapf("[ERROR] Error creating SSM parameter: {{err}}", err)
	}

	if err := setTagsSSM(ssmconn, d, d.Get("name").(string), "Parameter"); err != nil {
		return errwrap.Wrapf("[ERROR] Error creating SSM parameter tags: {{err}}", err)
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsSsmParameterRead(d, meta)
}
