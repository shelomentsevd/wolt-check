# Scripts and tools to extract data from Wolt receipts

This project helps extract and analyze receipt data from Wolt food delivery service.

## Prerequisites

- Google account with Wolt receipts in Gmail
- Claude API key
- Docker and Docker Compose
- Go 1.23.8

## How to use

1. Configure Gmail extraction:
    - Open scripts.google.com
    - Create new project and paste content from `scripts/extract_from_gmail.gs`
    - Run the script to extract PDFs to Google Drive

2. Download receipt PDFs:
    - Go to Google Drive
    - Download the extracted PDFs to `./data/pdfs` directory

3. Set up environment:
    - Copy `.env.example` to `.env`
    - Add your Claude API key to `.env`
    - Run `make dev-setup` to install required tools
    - Run `make docker-up` to start PostgreSQL
    - Run `make migrate-up` to initialize database schema

4. Process receipts:
    - Run `python scripts/parse_pdfs_by_ai.py` to extract data
    - Run `make build` to build the application
    - Run `make run` to process and store the data

## Development

- Use `make migrate-create name=migration_name` to create new migrations
- Use `make sqlc` to regenerate database code
- Use `make migrate-down` to rollback migrations