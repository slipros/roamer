# Roamer Documentation

This directory contains the documentation site for Roamer, built with Jekyll and deployed to GitHub Pages.

## Development

To run the documentation site locally:

1. Install Ruby 3.2 or later
2. Install dependencies:
   ```bash
   cd docs
   bundle install
   ```

3. Serve the site:
   ```bash
   bundle exec jekyll serve
   ```

4. Open http://localhost:4000/roamer in your browser

## Structure

- `index.md` - Homepage
- `getting-started.md` - Installation and basic usage
- `examples.md` - Comprehensive examples
- `api-reference.md` - Complete API documentation
- `extending.md` - Guide for creating custom components
- `_config.yml` - Jekyll configuration
- `_includes/` - Custom HTML includes
- `_layouts/` - Custom page layouts

## Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the main branch. The deployment is handled by the `.github/workflows/docs.yml` GitHub Actions workflow.

## Custom Domain

The documentation is served from `roamer.slipros.dev` as configured in the `CNAME` file.

## Features

- Responsive design with the Minima theme
- Syntax highlighting for code blocks
- Mermaid.js diagrams support
- Copy buttons for code blocks
- Anchor links for headings
- SEO optimization
- Mobile-friendly navigation