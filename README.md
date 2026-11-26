# lgtm-generator
You set input image then output LGTM image resized specific width.
Static images (jpg, png, webp, etc.) are always output as lossy WebP (quality 80). GIFs are output as animated GIFs.

## How to use
make output dir.
```sh
mkdir output
```

in local.
```sh
# Specify a file name in input/. The extension is optional.
go run . --input input-image
```

### Args
```sh
--input
      file name in input/. the extension can be omitted
      (if several files share the name, specify the extension)
--width
      target width. you can choose 320px, 500px
```

### Default
- output path: ./output/${width}_lgtm_${input-image-name}.webp
  (GIF: ./output/${width}_lgtm_${input-image-name}.gif)
- lgtm image path: ./assets/320_lgtm.png
- target width: 320px
