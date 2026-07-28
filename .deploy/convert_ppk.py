"""Convert PuTTY PPK v2 (no passphrase) to OpenSSH PEM."""
import base64, struct, sys, os

def read_ppk(path):
    with open(path, 'r') as f:
        lines = f.read().strip().split('\n')
    data = {}
    i = 0
    while i < len(lines):
        line = lines[i]
        if ': ' in line:
            key, val = line.split(': ', 1)
            if key.endswith('-Lines'):
                count = int(val)
                data[key.replace('-Lines', '')] = ''.join(lines[i+1:i+1+count])
                i += count + 1
                continue
            else:
                data[key] = val
        i += 1
    return data

def ppk_to_openssh_pem(ppk_path, pem_path):
    ppk = read_ppk(ppk_path)
    if ppk.get('Encryption', 'none') != 'none':
        print("ERROR: PPK is encrypted, need passphrase")
        sys.exit(1)

    pub_blob = base64.b64decode(ppk['Public'])
    priv_blob = base64.b64decode(ppk['Private'])

    # Build OpenSSH traditional RSA PEM (PKCS#1) from the raw components
    # PPK public: key_type_len + key_type + e_len + e + n_len + n
    # PPK private: d_len + d + p_len + p + q_len + q + iqmp_len + iqmp

    def read_mpint(data, offset):
        length = struct.unpack('>I', data[offset:offset+4])[0]
        return data[offset+4:offset+4+length], offset+4+length

    off = 0
    key_type_raw, off = read_mpint(pub_blob, off)
    e_raw, off = read_mpint(pub_blob, off)
    n_raw, off = read_mpint(pub_blob, off)

    off = 0
    d_raw, off = read_mpint(priv_blob, off)
    p_raw, off = read_mpint(priv_blob, off)
    q_raw, off = read_mpint(priv_blob, off)
    iqmp_raw, off = read_mpint(priv_blob, off)

    # Compute dp = d mod (p-1), dq = d mod (q-1)
    n = int.from_bytes(n_raw, 'big')
    e = int.from_bytes(e_raw, 'big')
    d = int.from_bytes(d_raw, 'big')
    p = int.from_bytes(p_raw, 'big')
    q = int.from_bytes(q_raw, 'big')
    iqmp = int.from_bytes(iqmp_raw, 'big')
    dp = d % (p - 1)
    dq = d % (q - 1)

    def int_to_der_bytes(v):
        b = v.to_bytes((v.bit_length() + 8) // 8, 'big')  # +8 to ensure leading 0 for positive
        return b

    def der_integer(b):
        # Ensure positive (add leading 0 if high bit set)
        if b[0] & 0x80:
            b = b'\x00' + b
        return der_tag(0x02, b)

    def der_tag(tag, content):
        l = len(content)
        if l < 0x80:
            return bytes([tag, l]) + content
        elif l < 0x100:
            return bytes([tag, 0x81, l]) + content
        else:
            return bytes([tag, 0x82, (l >> 8) & 0xff, l & 0xff]) + content

    # RSA PKCS#1 DER: SEQUENCE { version=0, n, e, d, p, q, dp, dq, iqmp }
    seq_content = (
        der_integer(b'\x00') +  # version
        der_integer(int_to_der_bytes(n)) +
        der_integer(int_to_der_bytes(e)) +
        der_integer(int_to_der_bytes(d)) +
        der_integer(int_to_der_bytes(p)) +
        der_integer(int_to_der_bytes(q)) +
        der_integer(int_to_der_bytes(dp)) +
        der_integer(int_to_der_bytes(dq)) +
        der_integer(int_to_der_bytes(iqmp))
    )
    der = der_tag(0x30, seq_content)
    b64 = base64.b64encode(der).decode()

    pem_lines = ['-----BEGIN RSA PRIVATE KEY-----']
    for i in range(0, len(b64), 64):
        pem_lines.append(b64[i:i+64])
    pem_lines.append('-----END RSA PRIVATE KEY-----')
    pem_lines.append('')

    with open(pem_path, 'w', newline='\n') as f:
        f.write('\n'.join(pem_lines))

    print(f"OK: {pem_path} ({os.path.getsize(pem_path)} bytes)")

if __name__ == '__main__':
    ppk_to_openssh_pem(sys.argv[1], sys.argv[2])
