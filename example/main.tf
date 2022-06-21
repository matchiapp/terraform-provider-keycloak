terraform {
  required_providers {
    keycloak = {
      source  = "terraform.local/mrparkers/keycloak"
      version = ">= 3.0"
    }
  }
}

provider "keycloak" {
  client_id     = "terraform"
  client_secret = "884e0f95-0f42-4a63-9b1f-94274655669e"
  url           = "http://localhost:8080"
  additional_headers = {
    foo = "bar"
  }
}

resource "keycloak_realm" "test" {
  realm             = "test"
  enabled           = true
  display_name      = "foo"
  display_name_html = "<b>foo</b>"

  smtp_server {
    host                  = "mysmtphost.com"
    port                  = 25
    from_display_name     = "Tom"
    from                  = "tom@myhost.com"
    reply_to_display_name = "Tom"
    reply_to              = "tom@myhost.com"
    ssl                   = true
    starttls              = true
    envelope_from         = "nottom@myhost.com"

    auth {
      username = "tom"
      password = "tom"
    }
  }

  account_theme        = "base"
  access_code_lifespan = "30m"

  internationalization {
    supported_locales = [
      "en",
      "de",
      "es",
    ]

    default_locale = "en"
  }

  security_defenses {
    headers {
      x_frame_options                     = "DENY"
      content_security_policy             = "frame-src 'self'; frame-ancestors 'self'; object-src 'none';"
      content_security_policy_report_only = ""
      x_content_type_options              = "nosniff"
      x_robots_tag                        = "none"
      x_xss_protection                    = "1; mode=block"
      strict_transport_security           = "max-age=31536000; includeSubDomains"
    }

    brute_force_detection {
      permanent_lockout                = false
      max_login_failures               = 31
      wait_increment_seconds           = 61
      quick_login_check_milli_seconds  = 1000
      minimum_quick_login_wait_seconds = 120
      max_failure_wait_seconds         = 900
      failure_reset_time_seconds       = 43200
    }
  }

  ssl_required    = "external"
  password_policy = "upperCase(1) and length(8) and forceExpiredPasswordChange(365) and notUsername"

  attributes = {
    mycustomAttribute  = "myCustomValue"
    userProfileEnabled = true
  }

  web_authn_policy {
    relying_party_entity_name = "Example"
    relying_party_id          = "keycloak.example.com"
    signature_algorithms = [
      "ES256",
      "RS256"
    ]
  }

  web_authn_passwordless_policy {
    relying_party_entity_name = "Example"
    relying_party_id          = "keycloak.example.com"
    signature_algorithms = [
      "ES256",
      "RS256"
    ]
  }
}


resource "keycloak_realm_user_profile" "userprofile" {
  realm_id = keycloak_realm.test.id

  attribute {
    name         = "field1"
    display_name = "Field 1"
    group        = "group1"

    enabled_when_scope = ["offline_access"]

    required_for_roles  = ["user"]
    required_for_scopes = ["offline_access"]

    permissions {
      view = ["admin", "user"]
      edit = ["admin", "user"]
    }

    validator {
      person_name_prohibited_characters {}
      pattern {
        pattern       = "^[a-z]+$"
        error_message = "Nope"
      }
      length {
        min = 5
        max = 10
      }

      double {
        max = 5.5
        min = 1.5
      }

    }

    annotations = {
      foo = "bar"
    }
  }

  attribute {
    name = "field2"
  }

  attribute {
    name         = "field3"
    display_name = "Field 3"

    validator {
      options {
        options = ["option1", "option2"]
      }
      email {}
      double {
        max = 5.5
        min = 1.5
      }
    }
  }

  group {
    name                = "group1"
    display_header      = "Group 1"
    display_description = "A first group"

    annotations = {
      foo = "bar"
    }
  }

  group {
    name = "group2"
  }
}
