# lgtm-generator
You set input image then output LGTM image resized specific width.

## How to use
make output dir.
```sh
mkdir output
```

in local.
```sh
# Specify relative path
go run . --input input-image
```

### Args
```sh
--input input path in input/
--width target width
```

### Default
- output path: ./output/${width}_lgtm_${input-image}
- lgtm path: ./assets/500_lgtm.png
- target width: 500px
