# route53-to-tf

Convert the result of `aws route53 list-resource-record-sets` to Terraform records.

## Usage

```bash
aws route53 list-resource-record-sets --hosted-zone-id <ZONE_ID> > records.json
cat records.json | go run main.go <DOMAIN>
```
