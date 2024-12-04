from io import StringIO
from bs4 import BeautifulSoup
import pandas as pd
import re
import requests

url = "https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-support"
response = requests.get(url)
html = response.content

soup = BeautifulSoup(html, "html.parser")

tables = soup.find_all("table")

dfs = []

for table in tables:
    parent_div = table.find_parent("div", class_="mx-tableFixed")

    # Find the h2 within the parent div
    title_element = parent_div.find_previous_sibling("h2")
    title = title_element.text.strip().lower().replace(" ", "")

    df = pd.read_html(StringIO(str(table)))[0]

    df["Resource type"] = df["Resource type"].apply(
        lambda x: (
            f"{title}/{x.lower().replace(' ', '')}"
            if title
            else x.lower().replace(" ", "")
        )
    )

    dfs.append(df)

df = pd.concat(dfs)

df = df[df["Supports tags"] == "Yes"]

df = df[["Resource type"]]

df.to_csv("azure_resource_tag_support.csv", index=False)
