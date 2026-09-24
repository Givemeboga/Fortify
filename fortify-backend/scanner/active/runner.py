from scanner.active.sqli import scan_sqli, scan_sqli_boolean
from scanner.active.xss import scan_xss
from scanner.active.path_traversal import scan_path_traversal

def run_active_scan(url: str) -> dict:
    return {
        "sqli": scan_sqli(url),
        "sqli_boolean": scan_sqli_boolean(url),
        "xss": scan_xss(url),
        "path_traversal": scan_path_traversal(url),
    }