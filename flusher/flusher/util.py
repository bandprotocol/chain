import re


# convert_proto_str_to_object convert input string to object
def convert_proto_str_to_object(s: str) -> object:
    pattern = r'(\w+):(?:([^\s"\']+)|["\']([^"\']*)["\'])'
    matches = re.findall(pattern, s)

    result = {}
    for key, unquoted_value, quoted_value in matches:
        result[key] = quoted_value if quoted_value else unquoted_value 
    return result



