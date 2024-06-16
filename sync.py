import boto3
import os
from botocore.exceptions import NoCredentialsError
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

endpoint_url = os.getenv('S3_ENDPOINT')
s3 = boto3.client(
    's3',
    endpoint_url=endpoint_url,
    aws_access_key_id=os.getenv('AWS_ACCESS_KEY_ID'),
    aws_secret_access_key=os.getenv('AWS_SECRET_ACCESS_KEY')
)

print(f"Using endpoint: {endpoint_url}")

def list_s3_files(bucket_name):
    s3_files = set()
    paginator = s3.get_paginator('list_objects_v2')
    for page in paginator.paginate(Bucket=bucket_name):
        for obj in page.get('Contents', []):
            s3_files.add(obj['Key'])
    return s3_files

def list_local_files(directory):
    local_files = set()
    for root, dirs, files in os.walk(directory):
        for file in files:
            local_path = os.path.join(root, file)
            relative_path = os.path.relpath(local_path, directory).replace("\\", "/")  # For Windows compatibility
            local_files.add(relative_path)
    return local_files

def upload_files(directory, bucket_name):
    local_files = list_local_files(directory)
    s3_files = list_s3_files(bucket_name)
    
    files_to_upload = local_files - s3_files
    print(f"Files to upload: {files_to_upload}")

    for relative_path in files_to_upload:
        local_path = os.path.join(directory, relative_path)
        try:
            s3.upload_file(local_path, bucket_name, relative_path)
            print(f"Uploaded {local_path} to {relative_path}")
        except FileNotFoundError:
            print(f"The file was not found: {local_path}")
        except NoCredentialsError:
            print("Credentials not available")

def download_files(bucket_name, directory):
    local_files = list_local_files(directory)
    s3_files = list_s3_files(bucket_name)
    
    files_to_download = s3_files - local_files
    print(f"Files to download: {files_to_download}")

    for s3_path in files_to_download:
        local_path = os.path.join(directory, s3_path)
        if not os.path.exists(os.path.dirname(local_path)):
            print(f"Creating directory: {os.path.dirname(local_path)}")
            os.makedirs(os.path.dirname(local_path))
        try:
            s3.download_file(bucket_name, s3_path, local_path)
            print(f"Downloaded {s3_path} to {local_path}")
        except NoCredentialsError:
            print("Credentials not available")

if __name__ == "__main__":
    local_directory = os.getenv('LOCAL_DIRECTORY')
    bucket_name = os.getenv('BUCKET_NAME')
    upload_files(local_directory, bucket_name)
    download_files(bucket_name, local_directory)
