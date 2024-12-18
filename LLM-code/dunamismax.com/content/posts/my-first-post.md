+++
title = "Self Hosting Hugo on Ubuntu Server"
date = "2024-12-18T18:54:18Z"
draft = false
+++

Getting my Hugo-powered website running on my Ubuntu server was surprisingly straightforward. I started by installing Hugo, then used the `hugo new site` command to set up a basic site framework. After selecting a theme, I tweaked my `config.toml` file and created my first Markdown post.

Next, I generated the static pages with a simple `hugo` command. The final step was serving the site behind Nginx—no complicated web apps or extra dependencies needed. Once the reverse proxy and SSL were in place, I pointed my domain at this server. Now, my website is up, secure, and fully self-hosted.

This quick process showed me how easy it is to manage a blog without relying on heavy frameworks. With Hugo’s simplicity, it feels like I finally have full control over my content, from the command line all the way to a live, public site.
