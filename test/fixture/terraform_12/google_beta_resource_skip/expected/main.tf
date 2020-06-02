// This resource should be skipped, as tfschema won't find resources that are exclusive to google-beta
// Note the google provider must be present in addition to the google-beta one

provider "google-beta" {
  project     = "my-project-id"
  region      = "us-central1"
}

data "google_billing_account" "account" {
  provider = google-beta
  billing_account = "000000-0000000-0000000-000000"
}

resource "google_billing_budget" "budget" {
  provider = google-beta
  billing_account = data.google_billing_account.account.id
  display_name = "Example Billing Budget"
  amount {
    specified_amount {
      currency_code = "USD"
      units = "100000"
    }
  }
  threshold_rules {
    threshold_percent =  0.5
  }
}