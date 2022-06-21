package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/mrparkers/terraform-provider-keycloak/keycloak"
)

func TestAccKeycloakRealmUserProfile_featureDisabled(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakRealmUserProfile_featureDisabled(realmName),
				ExpectError: regexp.MustCompile("User Profile is disabled"),
			},
		},
	})
}

func TestAccKeycloakRealmUserProfile_basicEmpty(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	realmUserProfile := &keycloak.RealmUserProfile{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmUserProfile_template(realmName, realmUserProfile),
				Check:  testAccCheckKeycloakRealmUserProfileExists("keycloak_realm_user_profile.realm_user_profile"),
			},
		},
	})
}

func TestAccKeycloakRealmUserProfile_basicFull(t *testing.T) {

	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	realmUserProfile := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{Name: "attribute1"},
			{
				Name:        "attribute2",
				DisplayName: "attribute 2",
				Group:       "group",
				Selector:    &keycloak.RealmUserProfileSelector{Scopes: []string{"roles"}},
				Required: &keycloak.RealmUserProfileRequired{
					Roles:  []string{"user"},
					Scopes: []string{"offline_access"},
				},
				Permissions: &keycloak.RealmUserProfilePermissions{
					Edit: []string{"admin", "user"},
					View: []string{"admin", "user"},
				},
				Validations: &keycloak.RealmUserProfileValidationConfig{
					Pattern: &keycloak.RealmUserProfileValidationPattern{
						Pattern:      "^[a-z]+$",
						ErrorMessage: "Error!",
					},
					PersonNameProhibitedChars: &keycloak.RealmUserProfileValidationProhibited{
						ErrorMessage: "Error!",
					},
					Length: &keycloak.RealmUserProfileValidationLength{
						Min:          5,
						Max:          10,
						TrimDisabled: false,
					},
					Integer:                 &keycloak.RealmUserProfileValidationInteger{Min: 5, Max: 99},
					Double:                  &keycloak.RealmUserProfileValidationDouble{Min: 4.5, Max: 65.7},
					URI:                     &map[string]interface{}{},
					Email:                   &map[string]interface{}{},
					LocalDate:               nil,
					UsernameProhibitedChars: &keycloak.RealmUserProfileValidationProhibited{ErrorMessage: "Error!"},
				},

				Annotations: map[string]string{"foo": "bar"},
			},
		},
		Groups: []*keycloak.RealmUserProfileGroup{
			{Name: "group", DisplayDescription: "Description", DisplayHeader: "Header", Annotations: map[string]string{"foo": "bar"}},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmUserProfile_template(realmName, realmUserProfile),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", realmUserProfile,
				),
			},
		},
	})
}

func TestAccKeycloakRealmUserProfile_group(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	withoutGroup := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{Name: "attribute"},
		},
	}

	withGroup := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{Name: "attribute"},
		},
		Groups: []*keycloak.RealmUserProfileGroup{
			{Name: "group"},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutGroup),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutGroup,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withGroup),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withGroup,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutGroup),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutGroup,
				),
			},
		},
	})
}

func TestAccKeycloakRealmUserProfile_attributeValidator(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	withoutValidator := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
			},
		},
	}

	withInitialConfig := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Validations: &keycloak.RealmUserProfileValidationConfig{
					Length: &keycloak.RealmUserProfileValidationLength{
						Min: 5,
						Max: 10,
					},
				},
			},
		},
	}

	withNewConfig := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Validations: &keycloak.RealmUserProfileValidationConfig{
					Length: &keycloak.RealmUserProfileValidationLength{
						Min: 6,
						Max: 10,
					},
				},
			},
		},
	}

	withNewValidator := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Validations: &keycloak.RealmUserProfileValidationConfig{
					PersonNameProhibitedChars: &keycloak.RealmUserProfileValidationProhibited{},
					Length: &keycloak.RealmUserProfileValidationLength{
						Min: 6,
						Max: 10,
					},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutValidator),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutValidator,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withInitialConfig),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withInitialConfig,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withNewConfig),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withNewConfig,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withNewValidator),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withNewValidator,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withNewConfig),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withNewConfig,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutValidator),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutValidator,
				),
			},
		},
	})
}

func TestAccKeycloakRealmUserProfile_options(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	withInitialOptions := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Required: &keycloak.RealmUserProfileRequired{
					Roles: []string{"admin", "user"},
				},
				Validations: &keycloak.RealmUserProfileValidationConfig{
					Options: &keycloak.RealmUserProfileValidationOptions{
						Options: []string{"option1", "option3", "option4"},
					},
				},
			},
		},
	}

	withNewOptions := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Required: &keycloak.RealmUserProfileRequired{
					Roles: []string{"admin", "user"},
				},
				Validations: &keycloak.RealmUserProfileValidationConfig{
					Options: &keycloak.RealmUserProfileValidationOptions{
						Options: []string{"option1", "option2", "option3", "option4"},
					},
				},
			},
		},
	}

	withNewSmallerOptions := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Required: &keycloak.RealmUserProfileRequired{
					Roles: []string{"admin", "user"},
				},
				Validations: &keycloak.RealmUserProfileValidationConfig{
					Options: &keycloak.RealmUserProfileValidationOptions{
						Options: []string{"option1", "option6"},
					},
				},
			},
		},
	}

	withoutOptions := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Required: &keycloak.RealmUserProfileRequired{
					Roles: []string{"admin", "user"},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withInitialOptions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withInitialOptions,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withNewOptions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withNewOptions,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withNewSmallerOptions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withNewSmallerOptions,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withNewSmallerOptions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withNewSmallerOptions,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutOptions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutOptions,
				),
			},
		},
	})
}

func TestAccKeycloakRealmUserProfile_attributePermissions(t *testing.T) {
	skipIfVersionIsLessThanOrEqualTo(testCtx, t, keycloakClient, keycloak.Version_14)

	realmName := acctest.RandomWithPrefix("tf-acc")

	withoutPermissions := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
			},
		},
	}

	viewAttributeMissing := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Permissions: &keycloak.RealmUserProfilePermissions{
					Edit: []string{"admin", "user"},
				},
			},
		},
	}

	editAttributeMissing := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Permissions: &keycloak.RealmUserProfilePermissions{
					View: []string{"admin", "user"},
				},
			},
		},
	}

	bothAttributesMissing := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name:        "attribute",
				Permissions: &keycloak.RealmUserProfilePermissions{},
			},
		},
	}

	withRightPermissions := &keycloak.RealmUserProfile{
		Attributes: []*keycloak.RealmUserProfileAttribute{
			{
				Name: "attribute",
				Permissions: &keycloak.RealmUserProfilePermissions{
					Edit: []string{"admin", "user"},
					View: []string{"admin", "user"},
				},
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviderFactories,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckKeycloakRealmUserProfileDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutPermissions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutPermissions,
				),
			},
			{
				Config:      testKeycloakRealmUserProfile_template(realmName, viewAttributeMissing),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			{
				Config:      testKeycloakRealmUserProfile_template(realmName, editAttributeMissing),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			{
				Config:      testKeycloakRealmUserProfile_template(realmName, bothAttributesMissing),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withRightPermissions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withRightPermissions,
				),
			},
			{
				Config: testKeycloakRealmUserProfile_template(realmName, withoutPermissions),
				Check: testAccCheckKeycloakRealmUserProfileStateEqual(
					"keycloak_realm_user_profile.realm_user_profile", withoutPermissions,
				),
			},
		},
	})
}

func testKeycloakRealmUserProfile_featureDisabled(realm string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}
resource "keycloak_realm_user_profile" "realm_user_profile" {
	realm_id = keycloak_realm.realm.id
}
`, realm)
}

func testKeycloakRealmUserProfile_template(realm string, realmUserProfile *keycloak.RealmUserProfile) string {
	tmpl, err := template.New("").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(`
resource "keycloak_realm" "realm" {
	realm 	   = "{{ .realm }}"

	attributes = {
		userProfileEnabled  = true
	}
}

resource "keycloak_realm_user_profile" "realm_user_profile" {
	realm_id = keycloak_realm.realm.id

	{{- range $_, $attribute := .userProfile.Attributes }}
	attribute {
        name = "{{ $attribute.Name }}"
		{{- if $attribute.DisplayName }}
        display_name = "{{ $attribute.DisplayName }}"
		{{- end }}

		{{- if $attribute.Group }}
        group = "{{ $attribute.Group }}"
		{{- end }}

		{{- if $attribute.Selector }}
		{{- if $attribute.Selector.Scopes }}
        enabled_when_scope = ["{{ StringsJoin $attribute.Selector.Scopes "\", \"" }}"]
		{{- end }}
		{{- end }}

		{{- if $attribute.Required }}
		{{- if $attribute.Required.Roles }}
        required_for_roles = ["{{ StringsJoin $attribute.Required.Roles "\", \"" }}"]
		{{- end }}
		{{- end }}

		{{- if $attribute.Required }}
		{{- if $attribute.Required.Scopes }}
        required_for_scopes = ["{{ StringsJoin $attribute.Required.Scopes "\", \"" }}"]
		{{- end }}
		{{- end }}

		{{- if $attribute.Permissions }}
        permissions {
			{{- if $attribute.Permissions.View }}
            view = ["{{ StringsJoin $attribute.Permissions.View "\", \"" }}"]
			{{- end }}
			{{- if $attribute.Permissions.Edit }}
            edit = ["{{ StringsJoin $attribute.Permissions.Edit "\", \"" }}"]
			{{- end }}
        }
		{{- end }}

		{{- if $attribute.Validations }}
        validator {
			{{- if $attribute.Validations.Length }}
			length {
				min = {{ $attribute.Validations.Length.Min }}
				max = {{ $attribute.Validations.Length.Max }}
				trim_disabled = {{ $attribute.Validations.Length.TrimDisabled }}
			}
			{{- end }}
			
			{{- if $attribute.Validations.Integer }}
			integer {
				min = {{ $attribute.Validations.Integer.Min }}
				max = {{ $attribute.Validations.Integer.Max }}
			}
			{{- end }}

			{{- if $attribute.Validations.Double }}
			double {
				min = {{ $attribute.Validations.Double.Min }}
				max = {{ $attribute.Validations.Double.Max }}
			}
			{{- end }}

			{{- if $attribute.Validations.URI }}
			uri {}
			{{- end }}

            {{- if $attribute.Validations.Pattern }}
			pattern {
				pattern = "{{ $attribute.Validations.Pattern.Pattern }}"
				error_message = "{{ $attribute.Validations.Pattern.ErrorMessage }}"
			}
			{{- end }}

			{{- if $attribute.Validations.Email }}
			email {}
			{{- end }}

			{{- if $attribute.Validations.LocalDate }}
			local_date {}
			{{- end }}

			{{- if $attribute.Validations.PersonNameProhibitedChars }}
			person_name_prohibited_characters {
				{{- if $attribute.Validations.PersonNameProhibitedChars.ErrorMessage }}
				error_message = "{{ $attribute.Validations.PersonNameProhibitedChars.ErrorMessage }}"
				{{- end }}
			}
			{{- end }}

			{{- if $attribute.Validations.UsernameProhibitedChars }}
			username_prohibited_characters {
				{{- if $attribute.Validations.UsernameProhibitedChars.ErrorMessage }}
				error_message = "{{ $attribute.Validations.UsernameProhibitedChars.ErrorMessage }}"
				{{- end }}
			}
			{{- end }}

			{{- if $attribute.Validations.Options }}
			options {
				options = [
					{{- range $attribute.Validations.Options.Options }}
					"{{ . }}",
					{{- end }}
				]
			}
			{{- end }}

        }
		{{- end }}

		{{- if $attribute.Annotations }}
        annotations = {
            {{- range $key, $value := $attribute.Annotations }}
            {{ $key }} = "{{ $value }}"
            {{- end }}
        }
		{{- end }}
    }
	{{- end }}

	{{- range $_, $group := .userProfile.Groups }}
    group {
        name = "{{ $group.Name }}"

		{{- if $group.DisplayHeader }}
        display_header = "{{ $group.DisplayHeader }}"
		{{- end }}

		{{- if $group.DisplayDescription }}
        display_description = "{{ $group.DisplayDescription }}"
		{{- end }}

		{{- if $group.Annotations }}
        annotations = {
            {{- range $key, $value := $group.Annotations }}
            {{ $key }} = "{{ $value }}"
            {{- end }}
        }
		{{- end }}
    }
	{{- end }}
}
	`)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	var tmplBuf bytes.Buffer
	err = tmpl.Execute(&tmplBuf, map[string]interface{}{"realm": realm, "userProfile": realmUserProfile})
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return tmplBuf.String()
}

func testAccCheckKeycloakRealmUserProfileExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getRealmUserProfileFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakRealmUserProfileStateEqual(resourceName string, realmUserProfile *keycloak.RealmUserProfile) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		realmUserProfileFromState, err := getRealmUserProfileFromState(s, resourceName)
		if err != nil {
			return err
		}

		// s1 := strings.ReplaceAll(string(j1), "\\", "")
		// s1 = strings.ReplaceAll(s1, "\"\"", "\"")
		// s1 = strings.ReplaceAll(s1, "\"[", "[")
		// s1 = strings.ReplaceAll(s1, "]\"", "]")

		if !reflect.DeepEqual(realmUserProfile, realmUserProfileFromState) {
			j1, _ := json.Marshal(realmUserProfile)
			j2, _ := json.Marshal(realmUserProfileFromState)

			return fmt.Errorf("%v\nshould be equal to\n%v", string(j1), string(j2))
		}

		return nil
	}
}

func testAccCheckKeycloakRealmUserProfileDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_realm_user_profile" {
				continue
			}

			realm := rs.Primary.Attributes["realm_id"]

			realmUserProfile, _ := keycloakClient.GetRealmUserProfile(testCtx, realm)
			if realmUserProfile != nil {
				return fmt.Errorf("user profile for realm %s", realm)
			}
		}

		return nil
	}
}

func getRealmUserProfileFromState(s *terraform.State, resourceName string) (*keycloak.RealmUserProfile, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm_id"]

	realmUserProfile, err := keycloakClient.GetRealmUserProfile(testCtx, realm)
	if err != nil {
		return nil, fmt.Errorf("error getting realm user profile: %s", err)
	}

	return realmUserProfile, nil
}
