API_BASE = 'http://localhost:8081/api/v2'
FOUNTAIN_BASE = 'https://fountain.toolforge.org/api/editathons'
from requests import Session
from dotenv import load_dotenv
import sqlite3
import os
# Load environment variables from .env file
load_dotenv()
sess = Session()
sess.cookies.update({
    'c-auth' : os.getenv('Auth'),
    'X-Refresh-Token' : os.getenv('Refresh')
})

def import_fountain(code):
    """
    Import a fountain script into the database.
    """
    url = f'{FOUNTAIN_BASE}/{code}'
    response = sess.get(url)
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
    response = sess.post(url, json=payload)
    if response.status_code != 200:
        raise Exception(f'Error creating campaign: {response.status_code}')
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
    response = sess.post(url, json=data)
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
    response = sess.post(url, json=data)
    if response.status_code != 200:
        raise Exception(f'Error pausing round: {response.status_code}')
    return response.json()
def _import_from_fountain(round_id, code):
    """
    Import a fountain script into a round.
    """
    url = f'{API_BASE}/round/import/{round_id}/fountain'
    data = {
        "code": code
    }
    response = sess.post(url, json=data)
    if response.status_code != 200:
        raise Exception(f'Error importing fountain script: {response.status_code}')
    task_id = response.json()['data']['taskId']
    print(f'Import task ID: {task_id}')
    # Periodically check the task status
    import time
    while True:
        status_url = f'{API_BASE}/task/{task_id}'
        status_response = sess.get(status_url)
        if status_response.status_code != 200:
            raise Exception(f'Error checking task status: {status_response.status_code}')
        status_data = status_response.json()['data']
        print(f'Task status: {status_data["status"]}')
        if status_data['status'] != 'pending':
            break
        print('Import in progress...')
        time.sleep(1)
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
def get_jury_id(username):
    """
    Get the jury ID for a given username.
    """
    with sqlite3.connect('data.db') as conn:
        cursor = conn.cursor()
        cursor.execute('SELECT id FROM user WHERE username = ?', (username,))
        result = cursor.fetchone()
        if result:
            return result[0]

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
def get_campaign_id(code):
    mp ={
        'fnf2025-bn' : 157,
        'fnf2025-tcy' : 158,
        'as-feminism-and-folklore-2025' : 160,
        'fnf2025-bs' : 161,
        'fnf2025-ks' : 163,
        'fnf2025-ur' : 162,
    }
    return mp.get(code, None)
if __name__ == '__main__':
    # Example usage
    # project_id = 'wiki-loves-folklore-international'
    

    codes = [
        'fnf2025-bn',
        'as-feminism-and-folklore-2025',
        'fnf2025-ur',
        'fnf2025-bs',
        'fnf2025-ks',
        'fnf2025-tcy',
    ]
    for code in codes:
        print(f'Importing fountain script: {code}')
        # import_fountain_script(project_id, code)
        votes = import_jury_votes(code)
        with sqlite3.connect('data.db') as conn:
            cursor = conn.cursor()
            cursor.execute('''
                CREATE TEMP TABLE IF NOT EXISTS jury_vote_temp (
                campaign_id INTEGER,
                page_title TEXT NOT NULL,
                jury_id INTEGER,
                score INTEGER,
                note TEXT DEFAULT '',
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                PRIMARY KEY (campaign_id, page_title, jury_id)
            );
            ''')
            cursor.executemany('''INSERT INTO jury_vote_temp (campaign_id, page_title, jury_id, score, note) VALUES (:campaign_id, :page_title, :jury_id, :score, :note)''', votes)
            # select all from the temp table to verify
            # cursor.execute('SELECT * FROM jury_vote_temp')
            cursor.execute("INSERT OR IGNORE INTO jury_vote SELECT j.jury_id, s.id, j.campaign_id, j.created_at, j.score,  j.note FROM jury_vote_temp j JOIN submission s ON j.page_title = s.title AND j.campaign_id = s.campaign_id AND j.jury_id <> s.submitted_by_id")
            conn.commit()
    # You can also create a campaign using the imported data
    # create_campaign(project_id=project_id, language='en', name='Example Campaign', description='This is an example campaign.', start_date='2023-01-01', end_date='2023-12-31', coordinators=['User1', 'User2'])

