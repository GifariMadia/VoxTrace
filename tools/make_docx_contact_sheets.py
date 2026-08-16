import sys
from pathlib import Path

from PIL import Image, ImageDraw


ROOT = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("build/extractions/qa-pages")


def main():
    for folder in sorted(p for p in ROOT.iterdir() if p.is_dir()):
        pages = sorted(folder.glob("page-*.png"))
        if not pages:
            continue
        thumbs = []
        for index, path in enumerate(pages, 1):
            image = Image.open(path).convert("RGB")
            image.thumbnail((340, 440))
            tile = Image.new("RGB", (360, 475), "#d8d8d8")
            tile.paste(image, ((360 - image.width) // 2, 20))
            ImageDraw.Draw(tile).text((12, 450), f"Page {index}", fill="black")
            thumbs.append(tile)
        columns = 4
        rows = (len(thumbs) + columns - 1) // columns
        sheet = Image.new("RGB", (columns * 360, rows * 475), "#bcbcbc")
        for index, tile in enumerate(thumbs):
            sheet.paste(tile, ((index % columns) * 360, (index // columns) * 475))
        sheet.save(ROOT / f"{folder.name}-contact.png")
        print(ROOT / f"{folder.name}-contact.png")


if __name__ == "__main__":
    main()
