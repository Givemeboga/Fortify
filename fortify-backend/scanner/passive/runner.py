from urllib.parse import urlparse

from scanner.passive.tls import scan_tls
from scanner.passive.headers import scan_headers
from scanner.passive.status import scan_paths

def run_passive_scan(url: str) -> dict:
    hostname = urlparse(url).hostname

    tls_result = scan_tls(hostname)
    headers_result = scan_headers(url)
    status_result = scan_paths(url)
    return {
        "tls": tls_result,
        "headers": headers_result,
        "status": status_result
    }