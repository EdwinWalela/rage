# rage
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT) [![Go Report Card](https://goreportcard.com/badge/github.com/edwinwalela/rage)](https://goreportcard.com/report/github.com/edwinwalela/rage)  

A minimalistic load testing tool designed to help developers stress-test their infrastructure with ease

<center>
<img src="./docs/rage.png" width=600/>
</center>

# Features
rage supports passing configuration arguments as CLI flags or through YAML

```
./rage -url <endpoint> -method <HTTP-Method> -users <integer> -attempts <integer>
```

Using a YAML configuration file enables you to customize the request further

```yaml
target:
  url: https://api.mydomain.com/v1/checkout
  method: POST
  successcode: 201
load:
  users: 5
  attempts: 10
headers:
  Authorization: Bearer eyJhbGciOIjp7ImlkIjoxLCJyb2iJ.eyJhbGciOiJ
body:
  content-type: application/json
  payload:
    product_uuid: 1e89b3e4567-e89b-12d3 
    quantity: 3
    colors: 
      - green
      - black
      - white
    extras:
      address:
        apartment_number: 10
        note: drop-off at reception 
      coupon_code: RAGE23!   
```

The above configuration models the following request body:

```json
{
  "product_uuid": "1e89b3e4567-e89b-12d3",
  "quantity": 3,
  "colors": ["green","black","white"],
  "extras": {
    "address": {
      "apartment_number": 10,
      "note": "drop-off at reception"
    },
    "coupon_code":"RAGE23!"
  }
}
```

Run rage with the above YAML configurations

```
./rage -f config.yaml
```

## How it works

rage makes the requests concurrently using goroutines.

- `users` - Number of virtual users to spawn
- `attempts` - Number of requests each user makes

## Contributing

Any contribution, in the form of a suggestion, bug report or pull request, is welcomed