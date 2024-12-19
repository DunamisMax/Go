+++
title = "Hosting a Hugo blog on Ubuntu"
date = "2024-12-18T12:00:00Z"
draft = false
+++

Over the past few days, I’ve set up a Hugo-based blog on an Ubuntu server. From installing Hugo and picking a theme to customizing layouts, adding a contact form, and tweaking the favicon, it’s been a step-by-step learning process.

## Installing Hugo & Starting a New Site

After installing Hugo (Extended) via apt, I ran:

```bash
hugo new site dunamismax.com
```

This created the basic structure. Using:

```bash
hugo server -D
```

I previewed changes instantly.

## Selecting and Customizing a Theme

I chose [hugo-paper](https://github.com/nanxiaobei/hugo-paper) for its simplicity. Setting `theme = "hugo-paper"` in `hugo.toml` integrated it smoothly. I then customized fonts, added a local `avatar.png` and `favicon.ico` in `static/`, and modified templates (`baseof.html`, `list.html`) to suit my style.

## Content & Contact Page

- **Home & Blog:** A home `_index.md` and `content/blog/` posts gave structure.
- **Contact:** A `contact.md` with an HTML form (using Formspree) required `unsafe = true` in Goldmark to render raw HTML.

## Troubleshooting

- **Draft & Future Dates:** Setting `draft = false` and current dates ensured pages built correctly.
- **Raw HTML:** Enabling unsafe rendering displayed the contact form properly.
- **Deployment Nuances:** On `https://dunamismax.com/`, I cleared caches and adjusted URLs until everything loaded as expected.
- **Favicon & Avatar:** Ensuring the correct paths in `static/` made these resources appear properly.

## Deployment

A final:

```bash
hugo
```

generated the `public/` directory. Uploading it to the server and configuring Nginx yielded a fast, fully static site.

## Final Thoughts

This process taught me that building a Hugo blog involves iterative tweaks, careful placement of files, and simple configuration changes. Each small step—theme selection, HTML form integration, or static file organization—built my confidence. I’m eager to continue refining the site and sharing what I learn along the way.
