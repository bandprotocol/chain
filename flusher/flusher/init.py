import json

import click

from .db import metadata, tracking
from .cli import cli

from sqlalchemy import create_engine


@cli.command()
@click.argument("chain_id")
@click.argument("topic")
@click.argument("replay_topic")
@click.option(
    "--db",
    help="Database URI connection string.",
    default="localhost:5432/postgres",
    show_default=True,
)
def init(chain_id, topic, replay_topic, db):
    """Initialize database with empty tables and tracking info."""
    engine = create_engine("postgresql+psycopg2://" + db, echo=True)
    metadata.create_all(engine)
    engine.execute(
        tracking.insert(),
        {"chain_id": chain_id, "topic": topic, "replay_topic": replay_topic, "kafka_offset": -1, "replay_offset": -2},
    )
    engine.execute(
        """
CREATE VIEW delegations_view
AS
  SELECT Cast(shares AS DECIMAL) * Cast(tokens AS DECIMAL) / Cast(
                   delegator_shares AS DECIMAL)                            AS
            amount,
         Cast(shares AS DECIMAL) / Cast(delegator_shares AS DECIMAL) * 100 AS
         share_percentage,
         Cast(shares AS DECIMAL) * Cast(current_reward AS DECIMAL) / Cast(
         delegator_shares AS DECIMAL) + ( Cast(current_ratio AS DECIMAL) - Cast(
                                          last_ratio AS DECIMAL) ) *
                                        Cast(shares AS DECIMAL)            AS
            reward,
         validators.operator_address,
         moniker,
         accounts.address                                                  AS
            delegator_address,
         IDENTITY
  FROM   delegations
         JOIN validators
           ON delegations.validator_id = validators.id
         JOIN accounts
           ON accounts.id = delegations.delegator_id;
"""
    )
    engine.execute(
        """
CREATE view validator_last_100_votes
AS
  SELECT Count(*),
         consensus_address,
         voted
  FROM   (SELECT *
          FROM   validator_votes
          ORDER  BY block_height DESC
          LIMIT  30000) tt
  WHERE  block_height > (SELECT Max(height)
                         FROM   blocks) - 101
  GROUP  BY consensus_address,
            voted;
"""
    )
    engine.execute(
        """
CREATE view validator_last_250_votes
AS
  SELECT Count(*),
         consensus_address,
         voted
  FROM   (SELECT *
          FROM   validator_votes
          ORDER  BY block_height DESC
          LIMIT  30000) tt
  WHERE  block_height > (SELECT Max(height)
                         FROM   blocks) - 251
  GROUP  BY consensus_address,
            voted;
"""
    )
    engine.execute(
        """
CREATE view validator_last_1000_votes
AS
  SELECT Count(*),
         consensus_address,
         voted
  FROM   (SELECT *
          FROM   validator_votes
          ORDER  BY block_height DESC
          LIMIT  100001) tt
  WHERE  block_height > (SELECT Max(height)
                         FROM   blocks) - 1001
  GROUP  BY consensus_address,
            voted;
"""
    )
    engine.execute(
        """
CREATE view validator_last_10000_votes
AS
  SELECT Count(*),
         consensus_address,
         voted
  FROM   (SELECT *
          FROM   validator_votes
          ORDER  BY block_height DESC
          LIMIT  1000001) tt
  WHERE  block_height > (SELECT Max(height)
                         FROM   blocks) - 10000
  GROUP  BY consensus_address,
            voted;
        """
    )
    engine.execute(
        """
CREATE VIEW oracle_script_statistic_last_1_day
AS
  SELECT Avg(resolve_time - request_time) AS response_time,
         Count(*)                         AS count,
         oracle_scripts.id,
         resolve_status
  FROM   oracle_scripts
         join requests
           ON oracle_scripts.id = requests.oracle_script_id
  WHERE  requests.request_time >= CAST(EXTRACT(epoch FROM NOW()) AS INT) - 86400
  GROUP  BY oracle_scripts.id,
            requests.resolve_status;
        """
    )
    engine.execute(
        """
CREATE VIEW oracle_script_statistic_last_1_week
AS
  SELECT Avg(resolve_time - request_time) AS response_time,
         Count(*)                         AS count,
         oracle_scripts.id,
         resolve_status
  FROM   oracle_scripts
         join requests
           ON oracle_scripts.id = requests.oracle_script_id
  WHERE  requests.request_time >= CAST(EXTRACT(epoch FROM NOW()) AS INT) - 604800
  GROUP  BY oracle_scripts.id,
            requests.resolve_status;
"""
    )
    engine.execute(
        """
CREATE VIEW oracle_script_statistic_last_1_month
AS
  SELECT Avg(resolve_time - request_time) AS response_time,
         Count(*)                         AS count,
         oracle_scripts.id,
         resolve_status
  FROM   oracle_scripts
         join requests
           ON oracle_scripts.id = requests.oracle_script_id
  WHERE  requests.request_time >= CAST(EXTRACT(epoch FROM NOW()) AS INT) - 2592000
  GROUP  BY oracle_scripts.id,
            requests.resolve_status;
"""
    )
    engine.execute(
        """
    CREATE VIEW non_validator_vote_proposals_view AS
    SELECT delegations.validator_id,
    votes.proposal_id,
    SUM(votes."yes" * shares * tokens / delegator_shares) as yes_vote,
    SUM(votes.abstain * shares * tokens / delegator_shares) as abstain_vote,
    SUM(votes."no" * shares * tokens / delegator_shares) as no_vote,
    SUM(votes.no_with_veto * shares * tokens / delegator_shares) as no_with_veto_vote
   FROM delegations
     JOIN votes ON delegator_id = voter_id
     JOIN validators ON validator_id = validators.id AND votes.voter_id != account_id
  GROUP BY validator_id, votes.proposal_id;
    """
    )

    engine.execute(
        """
    CREATE VIEW validator_vote_proposals_view AS
    SELECT validators.id,
           proposal_id,
           votes."yes" * tokens as yes_vote,
           votes.abstain * tokens as abstain_vote,
           votes."no" * tokens as no_vote,
           votes.no_with_veto * tokens as no_with_veto_vote
    FROM votes
    JOIN accounts ON accounts.id = votes.voter_id
    JOIN validators ON accounts.id = validators.account_id;
    """
    )
    engine.execute(
        """
BEGIN;
CREATE TABLE validator_report_count (validator_id integer primary key, count integer);
-- establish initial count
INSERT INTO validator_report_count SELECT validator_id, count(*) AS COUNT FROM reports GROUP BY validator_id;
CREATE OR REPLACE FUNCTION adjust_count()
RETURNS TRIGGER AS
$$
   DECLARE
   BEGIN
   IF TG_OP = 'INSERT' THEN
      EXECUTE 'INSERT INTO validator_report_count(validator_id, count)
        VALUES(''' || NEW.validator_id || ''', 1)
        ON CONFLICT (validator_id) DO UPDATE
            SET count = validator_report_count.count + 1';
      RETURN NEW;
   ELSIF TG_OP = 'DELETE' THEN
      EXECUTE 'UPDATE validator_report_count set count=count -1 WHERE validator_id = ''' || OLD.validator_id || '''';
      RETURN OLD;
   END IF;
   END;
$$
LANGUAGE 'plpgsql';
CREATE TRIGGER validator_report_count_trigger BEFORE INSERT OR DELETE ON reports
  FOR EACH ROW EXECUTE PROCEDURE adjust_count();
COMMIT;
"""
    )
