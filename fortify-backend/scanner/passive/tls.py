import ssl
import socket
from datetime import datetime

def scan_tls(hostname: str) -> dict:
    try:
        context = ssl.create_default_context()

        with socket.create_connection((hostname, 443)) as sock:
            with context.wrap_socket(sock, server_hostname=hostname) as tls_sock:
                version = tls_sock.version()
                cert = tls_sock.getpeercert()
                cipher = tls_sock.cipher()[0]
                expiry_date = datetime.strptime(cert['notAfter'], '%b %d %H:%M:%S %Y %Z')
                is_expired = expiry_date < datetime.utcnow()
                is_valid = expiry_date > datetime.utcnow() and bool(cert)

        return {
            "tls_version": version,
                    "cert_expired": is_expired,
                    "cert_valid": is_valid,
                    "cipher_suite": cipher
                }

    except Exception as e:
        return{
            "tls_version": None,
            "cert_expired": None,
            "cert_valid": False,
            "cipher_suite": None,
        }