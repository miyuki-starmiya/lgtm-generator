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

### Multiple files
Pass a glob pattern (`*`, `?`, `[...]`) to process every matching file in input/.
Quote the pattern, or the shell expands it in the current directory instead of input/
(zsh stops with `no matches found`).
```sh
go run . --input 'rose0*.*'
go run . --input 'rose0*'      # the extension can be omitted here too
go run . --input '*.gif' --width 500
```
- A file that fails does not stop the rest. The summary `Done: N succeeded, M failed` is printed at the end, and the exit code is 1 if any file failed.
- Files that differ only by extension (e.g. `chisato01.jpg` and `chisato01.webp`) would write the same output, so the second one is skipped and counted as failed.

### Args
```sh
--input
      file name in input/. the extension can be omitted
      (if several files share the name, specify the extension)
      or a quoted glob pattern such as 'rose0*' to process multiple files
--width
      target width. you can choose 320px, 500px
```

### Default
- output path: ./output/${width}_lgtm_${input-image-name}.webp
  (GIF: ./output/${width}_lgtm_${input-image-name}.gif)
- lgtm image path: ./assets/320_lgtm.png
- target width: 320px
