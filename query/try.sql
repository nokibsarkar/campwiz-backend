UPDATE evaluations e
SET judge_id = 'j2gmoqo5nz8cj'
WHERE 
    e.judge_id = 'j2gmoqo5nz8cl'
AND
    e.submission_id NOT IN (
        SELECT submission_id 
        FROM evaluations 
        WHERE judge_id = 'j2gmoqo5nz8cj'
        AND round_id = e.round_id
    )
AND
    e.round_id = 'r2gmoqo4a19fk'
AND
    e.score IS NULL
AND
    e.evaluation_id IN (
        -- Select the evaluation IDs to transfer, limited by available unassigned count
        SELECT sub.evaluation_id FROM (
            SELECT 
                e2.evaluation_id,
                ROW_NUMBER() OVER (ORDER BY e2.evaluation_id) as rn
            FROM evaluations e2
            WHERE e2.judge_id = 'j2gmoqo5nz8cl'
            AND e2.submission_id NOT IN (
                SELECT submission_id 
                FROM evaluations 
                WHERE judge_id = 'j2gmoqo5nz8cj'
                AND round_id = e2.round_id
            )
            AND e2.round_id = 'r2gmoqo4a19fk'
            AND e2.score IS NULL
        ) sub
        WHERE sub.rn <= (
            -- Count how many unassigned evaluations are available
            SELECT COUNT(DISTINCT e1.submission_id)
            FROM evaluations e1
            WHERE e1.submission_id NOT IN (
                SELECT submission_id 
                FROM evaluations 
                WHERE judge_id = 'j2gmoqo5nz8cl'
                AND round_id = e1.round_id
            ) 
            AND e1.round_id = 'r2gmoqo4a19fk'
            AND e1.judge_id IS NULL
            AND e1.score IS NULL
        )
    );
UPDATE evaluations e
SET
    judge_id = 'j2gmoqo5nz8cl'
WHERE 
    e.evaluation_id IN (
        SELECT 
            MAX(e1.evaluation_id)
        FROM
            evaluations e1
        WHERE
            e1.judge_id IS NULL
        AND
            e1.submission_id NOT IN (
                SELECT
                    submission_id 
                FROM
                    evaluations
                WHERE
                    judge_id = 'j2gmoqo5nz8cl'
                AND
                    round_id = e1.round_id
            )
        AND
            e1.score IS NULL
        AND
            e1.round_id = 'r2gmoqo4a19fk'
        GROUP BY e1.submission_id
) ORDER BY RAND()
LIMIT 6;
update roles join (SELECT COUNT(*) AS TotalAssigned, SUM(IF(evaluated_at IS NOT NULL, 1, 0)) AS TotalEvaluated, judge_id as role_id FROM `evaluations` WHERE round_id = 'r2gmoqo4a19fk' AND `judge_id` IS NOT NULL GROUP BY judge_id) k using(role_id)  set roles.total_evaluated=k.TotalEvaluated , roles.total_assigned=k.TotalAssigned;
