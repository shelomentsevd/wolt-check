#!/usr/bin/env python3
import os
import sys
import base64
import argparse
from anthropic import Anthropic

def extract_csv_from_pdf(api_key: str, pdf_path: str, model: str = "claude-3-7-sonnet-20250219") -> str:
    # Load and encode the PDF
    with open(pdf_path, "rb") as f:
        pdf_bytes = f.read()
    pdf_b64 = base64.b64encode(pdf_bytes).decode("utf-8")

    client = Anthropic(api_key=api_key)
    # Build the message payload
    response = client.messages.create(
        model=model,
        max_tokens=20000,
        temperature=1,
        system=(
            "You are a data-extraction assistant. I will supply you with one or more PDF files, "
            "each containing a Wolts food-delivery check (receipt). Your job is to:\n\n"
            "1. Read the PDF content and identify these fields:\n"
            "   - Receipt ID or Order Number\n"
            "   - Seller\n"
            "   - Venue\n"
            "   - Date and time\n"
            "   - Order type (if present)\n"
            "   - Payment method (if present)\n"
            "   - Customer name (if present)\n"
            "   - Delivery address (if present)\n"
            "   - For each line-item:\n"
            "     • Item name/description\n"
            "     • Quantity\n"
            "     • Unit price\n"
            "     • Line-item total\n"
            "   - Subtotal\n"
            "   - Tax amount(s) (if any)\n"
            "   - Delivery fee\n"
            "   - Tip (if any)\n"
            "   - Grand total\n\n"
            "2. Produce a single CSV table where:\n"
            "   - Each line-item gets its own row\n"
            "   - The columns are:\n"
            "     ReceiptID, Seller, Venue, DateTime, CustomerName, PaymentMethod,  DeliveryAddress, ItemName, Quantity, UnitPrice, "
            "LineTotal, Subtotal, Tax, DeliveryFee, Tip, GrandTotal\n\n"
            "3. If a field is missing on a particular receipt, leave its CSV cell blank.\n"
            "4. Do not include any extra commentary or markdown—output only valid CSV, with a header row."
        ),
        messages=[
            {
                "role": "user",
                "content": [
                    {
                        "type": "document",
                        "source": {
                            "type": "base64",
                            "media_type": "application/pdf",
                            "data": pdf_b64
                        }
                    }
                ]
            }
        ]
    )
    return response.content

def main():
    parser = argparse.ArgumentParser(description="Extract receipt data from a Wolts PDF to CSV via Claude.")
    parser.add_argument("input_pdf", help="Path to the input PDF file.")
    parser.add_argument("output_csv", help="Path to write the extracted CSV.")
    parser.add_argument(
        "--api-key", "-k",
        default=os.environ.get("ANTHROPIC_API_KEY"),
        help="Your Anthropi­c API key (or set ANTHROPIC_API_KEY env var)."
    )
    args = parser.parse_args()

    if not args.api_key:
        print("Error: API key must be provided via --api-key or ANTHROPIC_API_KEY env var.", file=sys.stderr)
        sys.exit(1)

    try:
        csv_output = extract_csv_from_pdf(args.api_key, args.input_pdf)
    except Exception as e:
        print(f"Failed to extract CSV: {e}", file=sys.stderr)
        sys.exit(1)

    # Write to output file
    with open(args.output_csv, "w", encoding="utf-8") as f:
        raw = csv_output
        if isinstance(raw, list):
            pieces = []
            for block in raw:
                if hasattr(block, "text"):
                    pieces.append(block.text)
                else:
                    pieces.append(str(block))
            f.write("".join(pieces))
        else:
            f.write(str(raw))
    print(f"Saved extracted CSV to {args.output_csv}")


if __name__ == "__main__":
    main()

