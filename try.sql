UPDATE `evaluations` e
 JOIN (
    SELECT 
        `evaluation_id`
    FROM
        `evaluations` e2
    WHERE e2.submission_id NOT IN (
    SELECT 
        `submission_id` 
    FROM 
        `evaluations` 
    WHERE 
        `judge_id` = @my_judge_id
            AND 
        `round_id` = @round_id
    UNION
        SELECT 
            `submission_id`
        FROM
            `submissions` s
        WHERE
            `s`.`round_id` = @round_id
                AND 
                `s`.`submitted_by_id` = @my_user_id
    ) AND 
        e2.round_id = @round_id 
    AND (
        e2.judge_id IS NULL 
            OR
        e2.judge_id IN (@reassignable_judges)
    ) AND 
        e2.score IS NULL 
    AND
        e2.evaluated_at IS NULL
    AND
        (
            e2.distribution_task_id IS NULL 
                OR
            e2.distribution_task_id <> @task_id
    ) GROUP BY 
        e2.submission_id
    LIMIT @N
 ) AS `e2`
    ON e.evaluation_id = e2.evaluation_id
    SET 
    e.judge_id = @my_judge_id, 
    e.distribution_task_id = @task_id 
LIMIT @N;
