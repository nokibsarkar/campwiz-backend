API_BASE = 'http://localhost:8081/api/v2'
FOUNTAIN_BASE = 'https://fountain.toolforge.org/api/editathons'
from requests import Session
from dotenv import load_dotenv
import sqlite3
import os
# Load environment variables from .env file
load_dotenv()
campwiz_sess = Session()
campwiz_sess.cookies.update({
    'c-auth' : os.getenv('Auth'),
    'X-Refresh-Token' : os.getenv('Refresh')
})
campwiz_bot_sess = Session()
campwiz_bot_sess.headers.update({
    'User-Agent': 'CampWiz Bot/1.0 (https://campwiz.org/bot)',
    'Authorization': f'Bearer {os.getenv("CampWizBotToken")}'
})
def import_fountain(code):
    """
    Import a fountain script into the database.
    """
    url = f'{FOUNTAIN_BASE}/{code}'
    response = campwiz_sess.get(url)
    if response.status_code != 200:
        raise Exception(f'Error fetching fountain script: {response.status_code}')
    data = response.json()
    return data

def create_campaign(*, project_id, language, name, description, start_date, end_date, coordinators : list[str]):
    """
    Create a new campaign in the database.
    """
    url = f'{API_BASE}/campaign/'
    payload = {
        "campaignType": "wikipedia",
        "coordinators": coordinators,
        "description": description,
        "endDate": end_date,
        "image": "string",
        "isPublic": False,
        "language": language,
        "name": name,
        # "organizers": organizers,
        "projectId": project_id,
        "rules": "",
        "startDate": start_date,
        "status": "PAUSED"
    }
    response = campwiz_sess.post(url, json=payload)
    if response.status_code != 200:
        raise Exception(f'Error creating campaign: {response.status_code} : {response.text}')
    return response.json()
def create_round(*, campaign_id, name, description, start_date, end_date, jury):
    data = {
        "allowJuryToParticipate": True,
        "allowMultipleJudgement": False,
        "allowedMediaTypes": [
            "ARTICLE"
        ],
        "articleAllowCreations": True,
        "articleAllowExpansions": True,
        "articleMaximumSubmissionOfSameArticle": 0,
        "articleMinimumAddedBytes": 0,
        "articleMinimumAddedWords": 0,
        "articleMinimumTotalBytes": 0,
        "articleMinimumTotalWords": 0,
        "audioMinimumDurationMilliseconds": 0,
        "audioMinimumSizeBytes": 0,
        "campaignId": campaign_id,
        "description": description,
        "endDate": end_date,
        "imageMinimumResolution": 0,
        "imageMinimumSizeBytes": 0,
        "isOpen": True,
        "isPublicJury": False,
        "jury": jury,
        "name": name,
        "quorum": 0,
        "secretBallot": True,
        "serial": 0,
        "startDate": start_date,
        "type": "binary",
        "videoMinimumDurationMilliseconds": 0,
        "videoMinimumResolution": 0,
        "videoMinimumSizeBytes": 0
        }
    url = f'{API_BASE}/round/'
    response = campwiz_sess.post(url, json=data)
    if response.status_code != 200:
        raise Exception(f'Error creating round: {response.status_code}')
    return response.json()
def pause_round(round_id):
    """
    Pause a round in the database.
    """
    url = f'{API_BASE}/round/{round_id}/status'
    data = {
        "status": "PAUSED"
    }
    response = campwiz_sess.post(url, json=data)
    if response.status_code != 200:
        raise Exception(f'Error pausing round: {response.status_code}')
    return response.json()
def check_import_task_status(task_id):
    """ Check the status of an import task.
    """
    # Periodically check the task status
    import time
    status_url = f'{API_BASE}/task/{task_id}'
    while True:
        status_response = campwiz_sess.get(status_url)
        if status_response.status_code != 200:
            raise Exception(f'Error checking task status: {status_response.status_code}')
        status_data = status_response.json()['data']
        print(f'Task status: {status_data["status"]}')
        if status_data['status'] != 'pending':
            return status_data['status']
        print('Import in progress...')
        time.sleep(1)

def _import_from_fountain(round_id, code):
    """
    Import a fountain script into a round.
    """
    url = f'{API_BASE}/round/import/{round_id}/fountain'
    data = {
        "code": code
    }
    response = campwiz_sess.post(url, json=data)
    if response.status_code != 200:
        raise Exception(f'Error importing fountain script: {response.status_code}')
    task_id = response.json()['data']['taskId']
    print(f'Import task ID: {task_id}')
    import_status = check_import_task_status(task_id)
    print(f'Import task completed with status: {import_status}')
    
def import_fountain_script(project_id, code):
    """
    Import a fountain script into the database.
    """
    data = import_fountain(code)
    del data['articles']
    language = data['wiki']
    name = data['name']
    description = data['description']
    start_date = data['start']
    end_date = data['finish']
    jury = data.get('jury', [])
    rules = data.get('rules', [])
    for rule in rules:
        type = rule.get('type', 'general')
        if type == 'articleCreated':
            params = rule.get('params', {})
            start_date = params.get('after', start_date)
            end_date = params.get('before', end_date)
    campaign = create_campaign(
        project_id=project_id,
        language=language,
        name=f'[Fountain] {name}',
        description=f'Imported from Fountain script: {code}\n\n{description}',
        start_date=start_date,
        end_date=end_date,
        coordinators=jury
    )
    campaign_id = campaign['data']['campaignId']
    print(f'Campaign ID: {campaign_id}')
    round = create_round(
        campaign_id=campaign_id,
        name=f'[Fountain] {name}',
        description=f'Imported from Fountain script: {code}\n\n{description}',
        start_date=start_date,
        end_date=end_date,
        jury=jury
    )
    round_id = round['data']['roundId']
    print(f'Round ID: {round_id}')
    pause_round(round_id)
    _import_from_fountain(round_id, code)
def create_campaign_from_v1(project_id, campaign_id_in_v1, path: str):
    with sqlite3.connect(path) as conn:
        conn.row_factory = sqlite3.Row
        cursor = conn.cursor()

        campaign_raw = cursor.execute('SELECT * FROM campaign WHERE id = ?', (campaign_id_in_v1,)).fetchone()
        jury_raw = cursor.execute('SELECT username FROM jury WHERE campaign_id = ?', (campaign_id_in_v1,)).fetchall()
        jury = list(map(lambda x: x['username'], jury_raw))
        start_at = campaign_raw['start_at'].replace(' ', 'T') + 'Z'
        end_at = campaign_raw['end_at'].replace(' ', 'T') + 'Z'
        campaign = create_campaign(
            project_id=project_id,
            language=campaign_raw['language'],
            name=campaign_raw['name'],
            description=f'Imported from v1 campaign ID: {campaign_id_in_v1}\n\n{campaign_raw["description"]}',
            start_date=start_at,
            end_date=end_at,
            coordinators=jury,
        )
        print(f'Created campaign: {campaign["data"]["name"]} with ID: {campaign["data"]["campaignId"]}')
        campaign_id = campaign['data']['campaignId']
        round = create_round(
            campaign_id=campaign_id,
            name=campaign_raw['name'],
            description=f'Imported from v1 campaign ID: {campaign_id_in_v1}\n\n{campaign_raw["description"]}',
            start_date=start_at,
            end_date=end_at,
            jury=jury
        )
        print(f'Created round: {round["data"]["name"]} with ID: {round["data"]["roundId"]}')
        return round['data']
def import_from_v1(project_id, campaign_id_in_v1, path: str, *, to_round_id=None):
    """
    Import a campaign from the v1 API.
    """
    if not to_round_id:
        round = create_campaign_from_v1(project_id, campaign_id_in_v1, path)
        to_round_id = round['roundId']
    url = f'{API_BASE}/round/import/{to_round_id}/campwizv1'
    data = {
        "fromCampaignId": campaign_id_in_v1,
        "fromFile": path,
        "toRoundId": to_round_id
    }
    response = campwiz_sess.post(url, json=data)
    if response.status_code != 200:
        raise Exception(f'Error importing campaign from v1: {response.status_code} : {response.text}')
    task_id = response.json()['data']['taskId']
    print(f'Import task ID: {task_id}')
    import_status = check_import_task_status(task_id)
    print(f'Import task completed with status: {import_status}')


def get_jury_id(username):
    """
    Get the jury ID for a given username.
    """
   

def import_jury_votes(code):
    campaign_id = get_campaign_id(code)
    if not campaign_id:
        raise Exception(f'Campaign not found for code: {code}')
    data = import_fountain(code)
    mark_policy = data.get('marks', {})

    articles = data.get('articles', [])
    votes = []
    for article in articles:
        page_title = article.get('name')
        marks = article.get('marks', {})
        for mark in marks:
            username = mark.get('user')
            if not username:
                continue
            jury_id = get_jury_id(username)
            if not jury_id:
                print(f'Jury ID not found for user: {username}')
                continue
            m = list(mark.get('marks').keys())[0]
            value = mark_policy[m]['values'][mark.get('marks')[m]]['value']
            votes.append({
                'campaign_id': campaign_id,
                'page_title': page_title,
                'jury_id': jury_id,
                'score': value,
                'note': mark.get('note', '')
            })
    return votes
mp ={
        'fnf2025-bn' : 157,
        'fnf2025-tcy' : 158,
        'as-feminism-and-folklore-2025' : 160,
        'fnf2025-bs' : 161,
        'fnf2025-ks' : 163,
        'fnf2025-ur' : 162,
    }
def get_campaign_id(code):
    return mp.get(code, None)
def resolve_redirects(language, page_titles):
    """
    Resolve redirects for a list of page titles in a given language.
    """
    url = f'https://{language}.wikipedia.org/w/api.php'
    normalization_map = {}
    redirect_map = {}
    # convert these pages in a batch of 50
    while page_titles:
        batch = page_titles[:50]
        page_titles = page_titles[50:]
        params = {
            "action": "query",
            "format": "json",
            "titles": '|'.join(batch),
            "redirects": 1,
            "formatversion": "2"
        }
        response = campwiz_bot_sess.get(url, params=params)
        if response.status_code != 200:
            raise Exception(f'Error fetching redirects: {response.status_code}')
        data = response.json()
        normalizations = data.get('query', {}).get('normalized', [])
        redirects = data.get('query', {}).get('redirects', [])
        for normalization in normalizations:
            normalization_map[normalization['to']] = normalization['from']
        for redirect in redirects:
            from_title = redirect['from']
            to_title = redirect['to']
            if from_title in normalization_map:
                redirect_map[normalization_map[from_title]] = normalization_map[to_title]
            else:
                redirect_map[from_title] = to_title
    return redirect_map

if __name__ == '__main__':
    # Example usage
    project_id = 'wiki-loves-folklore-international'
    

    codes = mp
    campaign_ids = [
        116,
        106,
        126,
        109,
        120,
        117,
        113,
        128,
        97,
        110,
        137,
        112,
        98,
        135,
        130,
        134,
        111,
        100,
        99,
        107,
        143,
        94,
        115,
        114,
        133,
        139,
        103,
        125,
        118,
        101,
        121,
        102,
        141,
        105,
        119
    ]

    UPDATE `submissions` JOIN (SELECT AVG(`evaluations`.`score`) As `Score`, COUNT(`evaluations`.`evaluation_id`) AS `AssignmentCount`, SUM(`evaluations`.`score` IS NOT NULL) AS `EvaluationCount`,`evaluations`.`submission_id`, evaluations.round_id as round_id FROM `evaluations` WHERE GROUP BY  evaluations.round_id, `evaluations`.`submission_id`) AS `e` ON `submissions`.`submission_id` = `e`.`submission_id` AND submissions.round_id=e.round_id SET `submissions`.`assignment_count` = `e`.`AssignmentCount`, `submissions`.`evaluation_count` = `e`.`EvaluationCount`, `submissions`.`score` = `e`.`Score` WHERE `submissions`.`round_id` = 'r2inkvjb95urk' 
    for campaign_id in campaign_ids:
        print(f'Importing campaign with ID: {campaign_id}')
        import_from_v1(project_id, campaign_id, 'data.db')
    for code in codes:
        print(f'Importing fountain script: {code}')
        campaign = codes.get(code, None)
        if not campaign:
            print(f'Campaign not found for code: {code}')
            continue
        import_fountain_script(project_id, code)
        # votes = import_jury_votes(code)
        # with sqlite3.connect('data.db') as conn:
        #     cursor = conn.cursor()
        #     cursor.execute('DROP TABLE IF EXISTS jury_vote_temp;')
        #     cursor.execute('''
        #         CREATE  TEMP TABLE IF NOT EXISTS jury_vote_temp (
        #         campaign_id INTEGER,
        #         page_title TEXT NOT NULL,
        #         jury_id INTEGER,
        #         score INTEGER,
        #         note TEXT DEFAULT '',
        #         created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        #         PRIMARY KEY (campaign_id, page_title, jury_id)
        #     );
        #     ''')
        #     j = cursor.executemany('''INSERT OR IGNORE INTO jury_vote_temp (campaign_id, page_title, jury_id, score, note) VALUES (:campaign_id, :page_title, :jury_id, :score, :note)''', votes)
        #     print(j.rowcount, 'votes inserted into jury_vote_temp')
        #     j = cursor.execute("INSERT OR IGNORE INTO jury_vote SELECT j.jury_id, s.id, s.campaign_id, j.created_at, j.score,  j.note FROM jury_vote_temp j LEFT JOIN submission s ON j.page_title = s.title AND j.campaign_id = s.campaign_id AND j.jury_id <> s.submitted_by_id")
        #     print(j.rowcount, 'votes inserted into jury_vote')
        #     if j.rowcount != len(votes):
        #         print('Some votes were not inserted due to conflicts or duplicates.')
        #         not_found_votes = "SELECT j.page_title FROM jury_vote_temp j LEFT JOIN submission s ON j.page_title = s.title AND j.campaign_id = s.campaign_id Where s.id IS NULL"
        #         cursor.execute(not_found_votes)
        #         missing_votes = [v[0] for v in cursor.fetchall()]
        #         redirect_map = resolve_redirects('as', missing_votes)
        #         newly_found_votes = []
        #         for vote in votes:
        #             if vote['page_title'] in redirect_map:
        #                 new_title = redirect_map[vote['page_title']]
        #                 newly_found_votes.append(vote)
        #                 r = cursor.execute('UPDATE jury_vote_temp SET page_title = ? WHERE campaign_id = ? AND jury_id = ? AND page_title = ?', (new_title, vote['campaign_id'], vote['jury_id'], vote['page_title']))
        #                 print('Updated', r.rowcount, 'votes with redirects for', vote['page_title'], 'to', new_title)
        #         print(f'Found {len(newly_found_votes)} votes that can be resolved with redirects.')
        #         j = cursor.executemany('''INSERT OR IGNORE INTO jury_vote SELECT j.jury_id, s.id, j.campaign_id, j.created_at, j.score,  j.note FROM jury_vote_temp j LEFT JOIN submission s ON j.page_title = s.title AND j.campaign_id = s.campaign_id AND j.jury_id <> s.submitted_by_id''', newly_found_votes)
        #         print(j.rowcount, 'votes inserted into jury_vote after resolving redirects')

        #     conn.commit()
    # You can also create a campaign using the imported data
    # create_campaign(project_id=project_id, language='en', name='Example Campaign', description='This is an example campaign.', start_date='2023-01-01', end_date='2023-12-31', coordinators=['User1', 'User2'])

