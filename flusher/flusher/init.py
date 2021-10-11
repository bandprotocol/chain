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
SELECT d.delegator_id,
       v.proposal_id,
       SUM(v."yes" * d.shares) AS yes_vote,
       SUM(v."abstain" * d.shares) AS abstain_vote,
       SUM(v."no" * d.shares) AS no_vote,
       SUM(v."no_with_veto" * d.shares) AS no_with_veto_vote
FROM delegations d
JOIN votes v ON d.delegator_id = v.voter_id
JOIN validators val ON d.validator_id = val.id
AND v.voter_id != val.account_id
GROUP BY d.delegator_id,
         v.proposal_id;
    """
    )

    engine.execute(
        """
CREATE VIEW validator_vote_proposals_view AS WITH non_val AS
  (SELECT v.proposal_id,
          val.account_id,
          SUM(CASE
                  WHEN v.voter_id != val.account_id THEN d.shares
                  ELSE 0
              END) AS Total
   FROM votes v
   JOIN delegations d ON d.delegator_id = v.voter_id
   JOIN validators val ON d.validator_id = val.id
   GROUP BY v.proposal_id,
            val.account_id)
SELECT v.proposal_id,
       val.account_id,
       v."yes" * (val.tokens - non_val.total) AS yes_vote,
       v."abstain" * (val.tokens - non_val.total) AS abstain_vote,
       v."no" * (tokens - non_val.total) AS no_vote,
       v."no_with_veto" * (tokens - non_val.total) AS no_with_veto_vote
FROM votes v
JOIN validators val ON val.account_id = v.voter_id
JOIN non_val ON non_val.proposal_id = v.proposal_id
AND non_val.account_id = val.account_id;
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

    engine.execute(
        """
CREATE VIEW proposal_total_votes AS
SELECT proposal_id,
       sum(all_vote.yes_vote) + sum(all_vote.abstain_vote) + sum(all_vote.no_vote) + sum(all_vote.no_with_veto_vote) AS SUM
FROM
  (SELECT proposal_id,
          yes_vote,
          abstain_vote,
          no_vote,
          no_with_veto_vote
   FROM non_validator_vote_proposals_view
   UNION SELECT proposal_id,
                yes_vote,
                abstain_vote,
                no_vote,
                no_with_veto_vote
   FROM validator_vote_proposals_view) all_vote
GROUP BY proposal_id
"""
    )
