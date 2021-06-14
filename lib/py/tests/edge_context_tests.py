import struct
import unittest

from baseplate.testing.lib.secrets import FakeSecretsStore

from reddit_edgecontext import EdgeContextFactory
from reddit_edgecontext import InvalidAuthenticationToken
from reddit_edgecontext import NoAuthenticationError
from reddit_edgecontext import ValidatedAuthenticationToken


AUTH_TOKEN_PUBLIC_KEY = """\
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtzMnDEQPd75QZByogNlB
NY2auyr4sy8UNTDARs79Edq/Jw5tb7ub412mOB61mVrcuFZW6xfmCRt0ILgoaT66
Tp1RpuEfghD+e7bYZ+Q2pckC1ZaVPIVVf/ZcCZ0tKQHoD8EpyyFINKjCh516VrCx
KuOm2fALPB/xDwDBEdeVJlh5/3HHP2V35scdvDRkvr2qkcvhzoy0+7wUWFRZ2n6H
TFrxMHQoHg0tutAJEkjsMw9xfN7V07c952SHNRZvu80V5EEpnKw/iYKXUjCmoXm8
tpJv5kXH6XPgfvOirSbTfuo+0VGqVIx9gcomzJ0I5WfGTD22dAxDiRT7q7KZnNgt
TwIDAQAB
-----END PUBLIC KEY-----
"""


AUTH_TOKEN_PRIVATE_KEY = """\
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAtzMnDEQPd75QZByogNlBNY2auyr4sy8UNTDARs79Edq/Jw5t
b7ub412mOB61mVrcuFZW6xfmCRt0ILgoaT66Tp1RpuEfghD+e7bYZ+Q2pckC1ZaV
PIVVf/ZcCZ0tKQHoD8EpyyFINKjCh516VrCxKuOm2fALPB/xDwDBEdeVJlh5/3HH
P2V35scdvDRkvr2qkcvhzoy0+7wUWFRZ2n6HTFrxMHQoHg0tutAJEkjsMw9xfN7V
07c952SHNRZvu80V5EEpnKw/iYKXUjCmoXm8tpJv5kXH6XPgfvOirSbTfuo+0VGq
VIx9gcomzJ0I5WfGTD22dAxDiRT7q7KZnNgtTwIDAQABAoIBAEjzXJw8yfAD391N
sCfG23mozrGzdd9cYu6fYCGSMSIX+kBiNV/l1DigxRzZ1bSggv4Am5H4LUj5HrAc
xTLLoMLHUIlkYfiYLc4zzE7qiHxR5AgLc5fq/FX9Uz8c0Kvgr3j+kVbrBOa7QONh
6SJ2L8aFap6kQMULbyFrSnaBY0omCZVV1tgyHwH2glZqm+0YjW4K0QU9KdE9YxLm
LXo8kXv5gZrPLtRzKrKo8WVscLGxSI0+x/ca9FjonGkZxUgfNZ61pu0sdXlpYCM/
yTjgtl35bFRcLsA1ZTxSI53y8iZUMljSg+oO6PEYrqzJh050khf/cP9GaSos00cJ
CUUx46kCgYEA8+GT8uIxLBTLQVvJMbvmK0enl1tvnQsSw6BKs1Gzprmk+7NRBeQg
6IuMAwMqgGgYq2MHPCt+/xEYlywyls37ME/JRlu0273lxK1jLcDAKgzHlCaMukzb
YyXdswRjkM0vzwljrjytUhO2Y8Bn8Y5LDR3ZP3XsZnx619NhHnWhPXMCgYEAwE2l
DNFXB6IZKzoBSrvG01yKxUjNoDDc6RaC1YciXJIDrya1EvJSsL0B7SOh2S2KWK4S
NdOMn3FlXGJU8CEwD67xDxWnRNFr26IMAOQAflxMiaooDBZevIY5wGBVUGU/45eo
fJK8+TmAbLMJWXT0PmY0HRI9Z9G8LLvbkTPqmbUCgYBQNz4mgWeTDDPeh4YhqaAH
VEY3bofDq8S8O5jWamUgzacKcuyPd32rf1rKEtyD57lVhfj3PYdD9ieUcaEiIRYh
ydx2G+vc6xUMH8ujXObX7eQnQpa/IFB4yBenriXhltvGyVvUD3kiqgEvxjVqKldd
NRgj23GqM/9jxc8H6cDjKQKBgBcOuEWLLDY8F2x5bgI59aMOK41wJ0eX1EUWb+WO
aG5VC6AKshdtesC6SOrYJOXXcgkkBgbyMKBFhnPHTovkTtDylvDKFd5Ihfg1u963
d9+uzXaMMnUQkZdlOLN6WYBkQ37Uhl05cvMnE+D0rqBNR6PbuU/rCfXSzg9HDxKf
+PcpAoGBAIu7BGSFmQpuL8OOcoap5moT2DHiatcDcsQGUg05kQxjWQPqkqUKk5Ms
Szq8xEe5u6rKLGFd1oPGB1/P6yiSxModVOGW0ALSlPNzHITtVEXLz27l3yXjBZeg
UNIHRjPq0YEdP4hwn2DxgZdgjm/RobXNz4DWfzRVqHR+hxMso5QQ
-----END RSA PRIVATE KEY-----
"""

AUTH_TOKEN_VALID = b"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoyNTI0NjA4MDAwfQ.dRzzfc9GmzyqfAbl6n_C55JJueraXk9pp3v0UYXw0ic6W_9RVa7aA1zJWm7slX9lbuYldwUtHvqaSsOpjF34uqr0-yMoRDVpIrbkwwJkNuAE8kbXGYFmXf3Ip25wMHtSXn64y2gJN8TtgAAnzjjGs9yzK9BhHILCDZTtmPbsUepxKmWTiEX2BdurUMZzinbcvcKY4Rb_Fl0pwsmBJFs7nmk5PvTyC6qivCd8ZmMc7dwL47mwy_7ouqdqKyUEdLoTEQ_psuy9REw57PRe00XCHaTSTRDCLmy4gAN6J0J056XoRHLfFcNbtzAmqmtJ_D9HGIIXPKq-KaggwK9I4qLX7g"  # noqa: E501
REQUEST_ID = "2adaff94-9067-4de0-a00b-79fded5cff9e"
SERIALIZED_EDGECONTEXT_WITH_NO_AUTH = b"\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0c\x00\x04\x0b\x00\x01\x00\x00\x00$becc50f6-ff3d-407a-aa49-fa49531363be\x00\x00\x0c\00\x07\00"  # noqa: E501
SERIALIZED_EDGECONTEXT_WITH_VALID_AUTH = b"\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0b\x00\x03\x00\x00\x01\xaeeyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoyNTI0NjA4MDAwfQ.dRzzfc9GmzyqfAbl6n_C55JJueraXk9pp3v0UYXw0ic6W_9RVa7aA1zJWm7slX9lbuYldwUtHvqaSsOpjF34uqr0-yMoRDVpIrbkwwJkNuAE8kbXGYFmXf3Ip25wMHtSXn64y2gJN8TtgAAnzjjGs9yzK9BhHILCDZTtmPbsUepxKmWTiEX2BdurUMZzinbcvcKY4Rb_Fl0pwsmBJFs7nmk5PvTyC6qivCd8ZmMc7dwL47mwy_7ouqdqKyUEdLoTEQ_psuy9REw57PRe00XCHaTSTRDCLmy4gAN6J0J056XoRHLfFcNbtzAmqmtJ_D9HGIIXPKq-KaggwK9I4qLX7g\x0c\x00\x04\x0b\x00\x01\x00\x00\x00$becc50f6-ff3d-407a-aa49-fa49531363be\x00\x0c\x00\x05\x0b\x00\x01\x00\x00\x00\tbaseplate\x00\x0c\x00\x06\x0b\x00\x01\x00\x00\x00\x02OK\x00\x0c\x00\x07\x00\x00"  # noqa: E501
SERIALIZED_EDGECONTEXT_WITH_EXPIRED_AUTH = b"\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0b\x00\x03\x00\x00\x01\xaeeyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0Ml9leGFtcGxlIiwiZXhwIjoxMjYyMzA0MDAwfQ.iUD0J2blW-HGtH86s66msBXymCRCgyxAZJ6xX2_SXD-kegm-KjOlIemMWFZtsNv9DJI147cNP81_gssewvUnhIHLVvXWCTOROasXbA9Yf2GUsjxoGSB7474ziPOZquAJKo8ikERlhOOVk3r4xZIIYCuc4vGZ7NfqFxjDGKAWj5Tt4VUiWXK1AdxQck24GyNOSXs677vIJnoD8EkgWqNuuwY-iFOAPVcoHmEuzhU_yUeQnY8D-VztJkip5-YPEnuuf-dTSmPbdm9ZTOP8gjTsG0Sdvb9NdLId0nEwawRy8CfFEGQulqHgd1bqTm25U-NyXQi7zroi1GEdykZ3w9fVNQ\x00\x0c\00\x07\00"  # noqa: E501
SERIALIZED_EDGECONTEXT_WITH_ANON_AUTH = b"\x0c\x00\x01\x0b\x00\x01\x00\x00\x00\x0bt2_deadbeef\n\x00\x02\x00\x00\x00\x00\x00\x01\x86\xa0\x00\x0c\x00\x02\x0b\x00\x01\x00\x00\x00\x08beefdead\x00\x0b\x00\x03\x00\x00\x01\xc0eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlcyI6WyJhbm9ueW1vdXMiXSwic3ViIjpudWxsLCJleHAiOjI1MjQ2MDgwMDB9.gQDiVzOUh70mKKK-YBTnLHWBOEuQyRllEE1-EIMfy3x5K8PsH9FB6Oy9S5HbILjfGFNrIBeux9HyW6hBDikoZDhn5QWyPNitL1pzMNONGGrXzSfaDoDbFy4MLD03A7zjG3qWBn_wLjgzUXX6qVX6W_gWO7dMqrq0iFvEegue-xQ1HGiXfPgnTrXRRovUO3JHy1LcZsmOjltYj5VGUTWXodBM8ObKEealDxg8yskEPy0IuujNMmb9eIyuHB8Ozzpg-lr790lxP37s5HCf18vrZ-IhRmLcLCqm5WSFyq_Ld2ByblBKL9pPst1AZYZTXNRIqovTAqr6v0-xjUeJ1iho9A\x00\x0c\00\x07\00"  # noqa: E501
SERIALIZED_EDGECONTEXT_WITH_NO_REQUEST_ID = (
    b"\x0c\x00\x01\x00\x0c\x00\x02\x00\x0c\x00\x04\x00\x0c\x00\x05\x00\x0c\x00\x06\x00" b"\x00"
)
SERIALIZED_EDGECONTEXT_WITH_READABLE_REQUEST_ID = (
    # Empty structs for the other fields
    b"\x0c\x00\x01\x00\x0c\x00\x02\x00\x0c\x00\x04\x00\x0c\x00\x05\x00\x0c\x00\x06\x00"
    # request_id
    + (
        # struct: request id
        b"\x0c\x00\x07"
        # string: readable_id
        b"\x0b\x00\x01"
        + struct.pack(">l", len(REQUEST_ID.encode("utf-8")))
        + REQUEST_ID.encode("utf-8")
        # end request id
        + b"\x00"
    )
    # end of EdgeContext
    + b"\x00"
)


class AuthenticationTokenTests(unittest.TestCase):
    def test_validated_authentication_token(self):
        payload = {
            "sub": "t2_user",
            "exp": 1574458470,
            "client_id": "client_id",
            "roles": ["role_a"],
            "client_type": "type_a",
            "scopes": ["scope_a"],
            "loid": {"id": "t2_user", "created_ms": 1574458470},
        }
        token = ValidatedAuthenticationToken(payload)
        self.assertEqual(token.subject, "t2_user")
        self.assertEqual(token.user_roles, {"role_a"})
        self.assertEqual(token.oauth_client_id, "client_id")
        self.assertEqual(token.oauth_client_type, "type_a")
        self.assertEqual(token.scopes, {"scope_a"})
        self.assertEqual(token.loid, "t2_user")
        self.assertEqual(token.loid_created_ms, 1574458470)

    def test_validated_authentication_token_none(self):
        payload = {
            "sub": "t2_user",
            "exp": 1574458470,
            "client_id": None,
            "client_type": None,
            "scopes": None,
            "loid": None,
        }
        token = ValidatedAuthenticationToken(payload)
        self.assertEqual(token.subject, "t2_user")
        self.assertEqual(token.user_roles, set())
        self.assertEqual(token.oauth_client_id, None)
        self.assertEqual(token.oauth_client_type, None)
        self.assertEqual(token.scopes, set())
        self.assertEqual(token.loid, None)
        self.assertEqual(token.loid_created_ms, None)

    def test_invalidated_authentication_token(self):
        token = InvalidAuthenticationToken()
        for attr in dir(token):
            if attr.startswith("__"):
                continue
            with self.assertRaises(NoAuthenticationError):
                getattr(token, attr)


class EdgeContextTests(unittest.TestCase):
    LOID_ID = "t2_deadbeef"
    LOID_CREATED_MS = 100000
    SESSION_ID = "beefdead"
    DEVICE_ID = "becc50f6-ff3d-407a-aa49-fa49531363be"
    ORIGIN_NAME = "baseplate"
    COUNTRY_CODE = "OK"

    def setUp(self):
        self.store = FakeSecretsStore(
            {
                "secrets": {
                    "secret/authentication/public-key": {
                        "type": "versioned",
                        "current": AUTH_TOKEN_PUBLIC_KEY,
                    }
                }
            }
        )
        self.factory = EdgeContextFactory(self.store)

    def test_create(self):
        request_context = self.factory.new(
            authentication_token=AUTH_TOKEN_VALID,
            loid_id=self.LOID_ID,
            loid_created_ms=self.LOID_CREATED_MS,
            session_id=self.SESSION_ID,
            device_id=self.DEVICE_ID,
            origin_service_name=self.ORIGIN_NAME,
            country_code=self.COUNTRY_CODE,
        )
        self.assertIsNot(request_context._t_request, None)
        self.assertEqual(request_context._header, SERIALIZED_EDGECONTEXT_WITH_VALID_AUTH)

    def test_create_validation(self):
        with self.assertRaises(ValueError):
            self.factory.new(
                authentication_token=None,
                loid_id="abc123",
                loid_created_ms=self.LOID_CREATED_MS,
                session_id=self.SESSION_ID,
            )
        with self.assertRaises(ValueError):
            self.factory.new(
                authentication_token=AUTH_TOKEN_VALID,
                loid_id=self.LOID_ID,
                loid_created_ms=self.LOID_CREATED_MS,
                session_id=self.SESSION_ID,
                country_code="aa",
            )

    def test_create_empty_context(self):
        request_context = self.factory.new()
        self.assertEqual(
            request_context._header,
            # loid
            b"\x0c\x00" b"\x01" b"\x00"  # STRUCT  # tag number  # END STRUCT
            # session
            b"\x0c\x00\x02\x00"
            # device
            b"\x0c\x00\x04\x00"
            # origin_service
            b"\x0c\x00\x05\x00"
            # geolocation
            b"\x0c\x00\x06\x00"
            # request_id
            b"\x0c\x00\x07\x00"
            # end of EdgeContext
            b"\x00",
        )

    def test_create_only_request_id(self):
        request_context = self.factory.new(request_id=REQUEST_ID)
        self.assertEqual(request_context._header, SERIALIZED_EDGECONTEXT_WITH_READABLE_REQUEST_ID)

    def test_logged_out_user(self):
        request_context = self.factory.from_upstream(SERIALIZED_EDGECONTEXT_WITH_NO_AUTH)

        with self.assertRaises(NoAuthenticationError):
            request_context.user.id
        with self.assertRaises(NoAuthenticationError):
            request_context.user.roles

        self.assertFalse(request_context.user.is_logged_in)
        self.assertEqual(request_context.user.loid, self.LOID_ID)
        self.assertEqual(request_context.user.cookie_created_ms, self.LOID_CREATED_MS)

        with self.assertRaises(NoAuthenticationError):
            request_context.oauth_client.id
        with self.assertRaises(NoAuthenticationError):
            request_context.oauth_client.is_type("third_party")

        self.assertEqual(request_context.session.id, self.SESSION_ID)
        self.assertEqual(request_context.device.id, self.DEVICE_ID)
        self.assertEqual(
            request_context.event_fields(),
            {
                "user_id": self.LOID_ID,
                "logged_in": False,
                "cookie_created_timestamp": self.LOID_CREATED_MS,
                "session_id": self.SESSION_ID,
                "oauth_client_id": None,
                "device_id": self.DEVICE_ID,
            },
        )

    def test_logged_in_user(self):
        request_context = self.factory.from_upstream(SERIALIZED_EDGECONTEXT_WITH_VALID_AUTH)

        self.assertEqual(request_context.user.id, "t2_example")
        self.assertTrue(request_context.user.is_logged_in)
        # For logged in user, we expect loid returns logged in user id
        self.assertEqual(request_context.user.loid, "t2_example")
        self.assertEqual(request_context.user.cookie_created_ms, self.LOID_CREATED_MS)
        self.assertEqual(request_context.user.roles, set())
        self.assertFalse(request_context.user.has_role("test"))
        self.assertIs(request_context.oauth_client.id, None)
        self.assertFalse(request_context.oauth_client.is_type("third_party"))
        self.assertEqual(request_context.session.id, self.SESSION_ID)
        self.assertEqual(request_context.device.id, self.DEVICE_ID)
        self.assertEqual(request_context.origin_service.name, self.ORIGIN_NAME)
        self.assertEqual(request_context.geolocation.country_code, self.COUNTRY_CODE)
        self.assertEqual(
            request_context.event_fields(),
            {
                "user_id": "t2_example",
                "logged_in": True,
                "cookie_created_timestamp": self.LOID_CREATED_MS,
                "session_id": self.SESSION_ID,
                "oauth_client_id": None,
                "device_id": self.DEVICE_ID,
            },
        )

    def test_expired_token(self):
        request_context = self.factory.from_upstream(SERIALIZED_EDGECONTEXT_WITH_EXPIRED_AUTH)

        with self.assertRaises(NoAuthenticationError):
            request_context.user.id
        with self.assertRaises(NoAuthenticationError):
            request_context.user.roles
        with self.assertRaises(NoAuthenticationError):
            request_context.oauth_client.id
        with self.assertRaises(NoAuthenticationError):
            request_context.oauth_client.is_type("third_party")

        self.assertFalse(request_context.user.is_logged_in)
        self.assertEqual(request_context.user.loid, self.LOID_ID)
        self.assertEqual(request_context.user.cookie_created_ms, self.LOID_CREATED_MS)
        self.assertEqual(request_context.session.id, self.SESSION_ID)
        self.assertEqual(
            request_context.event_fields(),
            {
                "user_id": self.LOID_ID,
                "logged_in": False,
                "cookie_created_timestamp": self.LOID_CREATED_MS,
                "session_id": self.SESSION_ID,
                "oauth_client_id": None,
            },
        )

    def test_anonymous_token(self):
        request_context = self.factory.from_upstream(SERIALIZED_EDGECONTEXT_WITH_ANON_AUTH)

        with self.assertRaises(NoAuthenticationError):
            request_context.user.id
        self.assertFalse(request_context.user.is_logged_in)
        self.assertEqual(request_context.user.loid, self.LOID_ID)
        self.assertEqual(request_context.user.cookie_created_ms, self.LOID_CREATED_MS)
        self.assertEqual(request_context.session.id, self.SESSION_ID)
        self.assertTrue(request_context.user.has_role("anonymous"))

    def test_request_id(self):
        request_context = self.factory.from_upstream(
            SERIALIZED_EDGECONTEXT_WITH_READABLE_REQUEST_ID
        )

        self.assertEqual(request_context.request_id.readable_id, REQUEST_ID)
        self.assertEqual(
            request_context.event_fields(),
            {
                "cookie_created_timestamp": None,
                "user_id": None,
                "logged_in": False,
                "session_id": None,
                "oauth_client_id": None,
                "edge_request_id": REQUEST_ID,
            },
        )
