# lgtm-generator
You set input image then output LGTM image resized specific width.

## How to use
in local
```sh
# Specify relative path
go run . --input input-image
```

### Default
- output path: $PWD/output/lgtm_${input-image}
- lgtm path: ./assets/lgtm.png
- resized width: 500px
