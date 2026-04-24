# Copyright IBM Corp. 2026
# SPDX-License-Identifier: MPL-2.0

# Reference: https://github.com/hashicorp/security-scanner/blob/main/CONFIG.md#binary (private repository)

repository {
  go_modules    = true
  osv           = true
  trojan_source = true

  secrets {
    all = true
  }

  github_actions {
    pinned_hashes = true
  }

  dockerfile {
    pinned_hashes = true
    curl_bash     = true
  }
}

binary {
  secrets {
    all = true
  }
  go_modules = true
  osv        = true
  nvd        = false
}
