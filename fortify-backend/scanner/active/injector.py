from urllib.parse import urlparse, parse_qs, urlencode, urlunparse

def inject_payload(url: str, payload:str) -> dict:
    parsed = urlparse(url)
    params = parse_qs(parsed.query)

    injected_urls = {}

    for param in params:
        modified = dict(params)
        modified[param] = payload
        new_query = urlencode(modified, doseq=True)
        new_url = urlunparse((parsed._replace(query=new_query)))
        injected_urls[param] = new_url
    return injected_urls