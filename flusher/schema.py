import requests
from flusher.db import metadata

# URL = "http://35.187.228.10:5433/v1"
# URL = "http://graphql-gm-lb.bandchain.org/v1"
# URL = "http://34.126.91.209/v1"
URL = "http://35.187.228.10/v1"
# VIEW_TABLE = [
#     "delegations_view",
#     "non_validator_vote_proposals_view",
#     "oracle_script_statistic_last_1_day",
#     "oracle_script_statistic_last_1_month",
#     "oracle_script_statistic_last_1_week",
#     "validator_last_10000_votes",
#     "validator_last_1000_votes",
#     "validator_last_100_votes",
#     "validator_last_250_votes",
#     "validator_vote_proposals_view",
# ]


# for table_name in metadata.tables.keys():
#     print(
#         requests.post(
#             URL + "/query",
#             headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
#             json={"type": "track_table", "args": {"schema": "public", "name": table_name},},
#         ).json()
#     )

# for table_name in VIEW_TABLE:
#     print(
#         requests.post(
#             URL + "/query",
#             headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
#             json={"type": "track_table", "args": {"schema": "public", "name": table_name},},
#         ).json()
#     )


print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_object_relationship",
            "args": {
                "table": "oracle_scripts",
                "name": "request_stat",
                "using": {
                    "manual_configuration": {
                        "remote_table": "oracle_script_requests",
                        "column_mapping": {"id": "oracle_script_id"},
                    }
                },
            },
        },
    ).json()
)

print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_array_relationship",
            "args": {
                "table": "oracle_scripts",
                "name": "response_last_1_day",
                "using": {
                    "manual_configuration": {
                        "remote_table": "oracle_script_statistic_last_1_day",
                        "column_mapping": {"id": "id"},
                    }
                },
            },
        },
    ).json()
)

print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_array_relationship",
            "args": {
                "table": "oracle_scripts",
                "name": "response_last_1_week",
                "using": {
                    "manual_configuration": {
                        "remote_table": "oracle_script_statistic_last_1_week",
                        "column_mapping": {"id": "id"},
                    }
                },
            },
        },
    ).json()
)

print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_array_relationship",
            "args": {
                "table": "oracle_scripts",
                "name": "response_last_1_month",
                "using": {
                    "manual_configuration": {
                        "remote_table": "oracle_script_statistic_last_1_month",
                        "column_mapping": {"id": "id"},
                    }
                },
            },
        },
    ).json()
)

print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_object_relationship",
            "args": {
                "table": "data_sources",
                "name": "request_stat",
                "using": {
                    "manual_configuration": {
                        "remote_table": "data_source_requests",
                        "column_mapping": {"id": "data_source_id"},
                    }
                },
            },
        },
    ).json()
)


# print(
#     requests.post(
#         URL + "/query",
#         headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
#         json={
#             "type": "drop_relationship",
#             "args": {"table": "redelegations", "relationship": "validator"},
#         },
#     ).json()
# )


print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_object_relationship",
            "args": {
                "table": "redelegations",
                "name": "validatorByValidatorDstId",
                "using": {
                    "manual_configuration": {
                        "remote_table": "validators",
                        "column_mapping": {"validator_dst_id": "id"},
                    }
                },
            },
        },
    ).json()
)

print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_object_relationship",
            "args": {
                "table": "historical_oracle_statuses",
                "name": "validator",
                "using": {
                    "manual_configuration": {
                        "remote_table": "validators",
                        "column_mapping": {"operator_address": "operator_address"},
                    }
                },
            },
        },
    ).json()
)

print(
    requests.post(
        URL + "/query",
        headers={"Content-Type": "application/json", "X-Hasura-Role": "admin"},
        json={
            "type": "create_object_relationship",
            "args": {
                "table": "accounts",
                "name": "validator",
                "using": {
                    "manual_configuration": {"remote_table": "validators", "column_mapping": {"id": "account_id"},}
                },
            },
        },
    ).json()
)

