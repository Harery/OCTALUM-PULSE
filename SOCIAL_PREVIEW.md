# Social Preview — OCTALUM-PULSE

GitHub social preview image spec. Owner uploads via **Settings → Social preview**.

## Required dimensions

- **1280 × 640 px** (2:1)
- PNG or JPG, < 1 MB
- Readable when scaled to ~640 × 320 (Twitter card / LinkedIn share)

## Layout

```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│   [HEART-PULSE GLYPH]   OCTALUM-PULSE                        │
│                                                              │
│       Your Infrastructure's Heartbeat                        │
│       One command. All distros. Zero worries.                │
│                                                              │
│       apt • dnf • pacman • zypper • apk                      │
│                                                              │
│                       github.com/Harery/OCTALUM-PULSE        │
└──────────────────────────────────────────────────────────────┘
```

## Color palette (from existing brand)

- Background: `#0A1628` (deep navy)
- Accent / pulse line: `#00D4FF` (electric cyan)
- Text primary: `#FFFFFF`
- Text muted: `#B0C4D9`

## Typography

- Headline: **Inter Bold** or **JetBrains Mono Bold**, ~96 px
- Tagline: **Inter Medium**, ~36 px
- Distro chips: **JetBrains Mono**, ~24 px

## Source asset

A pre-rendered SVG is already at `.github/social-preview.svg`. Either:

1. Open in Figma / Inkscape, export PNG at 1280×640, upload, OR
2. Use ImageMagick:
   ```bash
   rsvg-convert -w 1280 -h 640 .github/social-preview.svg \
     -o social-preview.png
   ```

## Verification

After upload, share the repo URL in a private channel and confirm the OG card renders correctly on:

- Twitter / X (`https://cards-dev.twitter.com/validator`)
- LinkedIn (`https://www.linkedin.com/post-inspector/`)
- Discord (paste link in a test channel)
