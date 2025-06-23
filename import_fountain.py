API_BASE = 'http://localhost:8081/api/v2'
FOUNTAIN_BASE = 'https://fountain.toolforge.org/api/editathons'
from requests import Session
from dotenv import load_dotenv
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

if __name__ == '__main__':
    # Example usage
    project_id = 'wiki-loves-folklore-international'
#     Tiven, [21.06.2025 22:59]
# https://fountain.toolforge.org/editathons/fnf2025-bn

# Tiven, [21.06.2025 22:59]
# https://fountain.toolforge.org/editathons/fnf2025-tcy

# Tiven, [21.06.2025 22:59]
# https://fountain.toolforge.org/editathons/fnf2025-tl

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/fnf2025-gu

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/as-feminism-and-folklore-2025

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/fnf2025-ml

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/fnf-2025-sd

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/fnf2025-ur

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/fnf2025-bs

# Tiven, [21.06.2025 23:00]
# https://fountain.toolforge.org/editathons/fnf2025-ks
    codes = [
        'fnf2025-bn',
        'fnf2025-tl',
        'as-feminism-and-folklore-2025',
        'fnf2025-ml',
        'fnf2025-ur',
        'fnf2025-bs',
        'fnf2025-ks',
        
        'fnf2025-tcy',
    ]
    for code in codes:
        print(f'Importing fountain script: {code}')
        import_fountain_script(project_id, code)
    # You can also create a campaign using the imported data
    # create_campaign(project_id=project_id, language='en', name='Example Campaign', description='This is an example campaign.', start_date='2023-01-01', end_date='2023-12-31', coordinators=['User1', 'User2'])
    