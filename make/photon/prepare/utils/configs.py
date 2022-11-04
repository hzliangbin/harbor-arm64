import os
import yaml
import logging
from g import versions_file_path, host_root_dir, DEFAULT_UID, INTERNAL_NO_PROXY_DN
from utils.misc import generate_random_string, owner_can_read, other_can_read

default_db_max_idle_conns = 2  # NOTE: https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns
default_db_max_open_conns = 0  # NOTE: https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns
default_https_cert_path = '/your/certificate/path'
default_https_key_path = '/your/certificate/path'


def validate(conf: dict, **kwargs):
    # hostname validate
    if conf.get('hostname') == '127.0.0.1':
        raise Exception("127.0.0.1 can not be the hostname")
    if conf.get('hostname') == 'reg.mydomain.com':
        raise Exception("Please specify hostname")

    # protocol validate
    protocol = conf.get("protocol")
    if protocol != "https" and kwargs.get('notary_mode'):
        raise Exception(
            "Error: the protocol must be https when Harbor is deployed with Notary")
    if protocol == "https":
        if not conf.get("cert_path") or conf["cert_path"] == default_https_cert_path:
            raise Exception("Error: The protocol is https but attribute ssl_cert is not set")
        if not conf.get("cert_key_path") or conf['cert_key_path'] == default_https_key_path:
            raise Exception("Error: The protocol is https but attribute ssl_cert_key is not set")
    if protocol == "http":
        logging.warning("WARNING: HTTP protocol is insecure. Harbor will deprecate http protocol in the future. Please make sure to upgrade to https")

    # log endpoint validate
    if ('log_ep_host' in conf) and not conf['log_ep_host']:
        raise Exception('Error: must set log endpoint host to enable external host')
    if ('log_ep_port' in conf) and not conf['log_ep_port']:
        raise Exception('Error: must set log endpoint port to enable external host')
    if ('log_ep_protocol' in conf) and (conf['log_ep_protocol'] not in ['udp', 'tcp']):
        raise Exception("Protocol in external log endpoint must be one of 'udp' or 'tcp' ")

    # Storage validate
    valid_storage_drivers = ["filesystem", "azure", "gcs", "s3", "swift", "oss"]
    storage_provider_name = conf.get("storage_provider_name")
    if storage_provider_name not in valid_storage_drivers:
        raise Exception("Error: storage driver %s is not supported, only the following ones are supported: %s" % (
            storage_provider_name, ",".join(valid_storage_drivers)))

    storage_provider_config = conf.get("storage_provider_config") ## original is registry_storage_provider_config
    if storage_provider_name != "filesystem":
        if storage_provider_config == "":
            raise Exception(
                "Error: no provider configurations are provided for provider %s" % storage_provider_name)
    # ca_bundle validate
    if conf.get('registry_custom_ca_bundle_path'):
        registry_custom_ca_bundle_path = conf.get('registry_custom_ca_bundle_path') or ''
        if registry_custom_ca_bundle_path.startswith('/data/'):
            ca_bundle_host_path = registry_custom_ca_bundle_path
        else:
            ca_bundle_host_path = os.path.join(host_root_dir, registry_custom_ca_bundle_path.lstrip('/'))
        try:
            uid = os.stat(ca_bundle_host_path).st_uid
            st_mode = os.stat(ca_bundle_host_path).st_mode
        except Exception as e:
            logging.error(e)
            raise Exception('Can not get file info')
        err_msg = 'Cert File {} should be owned by user with uid 10000 or readable by others'.format(registry_custom_ca_bundle_path)
        if uid == DEFAULT_UID and not owner_can_read(st_mode):
            raise Exception(err_msg)
        if uid != DEFAULT_UID and not other_can_read(st_mode):
            raise Exception(err_msg)

    # Redis validate
    redis_host = conf.get("redis_host")
    if redis_host is None or len(redis_host) < 1:
        raise Exception(
            "Error: redis_host in harbor.yml needs to point to an endpoint of Redis server or cluster.")

    redis_port = conf.get("redis_port")
    if redis_host is None or (redis_port < 1 or redis_port > 65535):
        raise Exception(
            "Error: redis_port in harbor.yml needs to point to the port of Redis server or cluster.")

    # TODO:
    # If user enable trust cert dir, need check if the files in this dir is readable.

def parse_versions():
    if not versions_file_path.is_file():
        return {}
    with open('versions') as f:
        versions = yaml.load(f)
    return versions

def parse_yaml_config(config_file_path, with_notary, with_clair, with_chartmuseum):
    '''
    :param configs: config_parser object
    :returns: dict of configs
    '''

    with open(config_file_path) as f:
        configs = yaml.load(f)

    config_dict = {
        'adminserver_url': "http://adminserver:8080",
        'registry_url': "http://registry:5000",
        'registry_controller_url': "http://registryctl:8080",
        'core_url': "http://core:8080",
        'core_local_url': "http://127.0.0.1:8080",
        'token_service_url': "http://core:8080/service/token",
        'jobservice_url': 'http://jobservice:8080',
        'clair_url': 'http://clair:6060',
        'clair_adapter_url': 'http://clair-adapter:8080',
        'notary_url': 'http://notary-server:4443',
        'chart_repository_url': 'http://chartmuseum:9999'
    }

    config_dict['hostname'] = configs["hostname"]

    config_dict['protocol'] = 'http'
    http_config = configs.get('http') or {}
    config_dict['http_port'] = http_config.get('port', 80)

    https_config = configs.get('https')
    if https_config:
        config_dict['protocol'] = 'https'
        config_dict['https_port'] = https_config.get('port', 443)
        config_dict['cert_path'] = https_config["certificate"]
        config_dict['cert_key_path'] = https_config["private_key"]

    if configs.get('external_url'):
        config_dict['public_url'] = configs.get('external_url')
    else:
        if config_dict['protocol'] == 'https':
            if config_dict['https_port'] == 443:
                config_dict['public_url'] = '{protocol}://{hostname}'.format(**config_dict)
            else:
                config_dict['public_url'] = '{protocol}://{hostname}:{https_port}'.format(**config_dict)
        else:
            if config_dict['http_port'] == 80:
                config_dict['public_url'] = '{protocol}://{hostname}'.format(**config_dict)
            else:
                config_dict['public_url'] = '{protocol}://{hostname}:{http_port}'.format(**config_dict)

    # DB configs
    db_configs = configs.get('database')
    if db_configs:
        # harbor db
        config_dict['harbor_db_host'] = 'postgresql'
        config_dict['harbor_db_port'] = 5432
        config_dict['harbor_db_name'] = 'registry'
        config_dict['harbor_db_username'] = 'postgres'
        config_dict['harbor_db_password'] = db_configs.get("password") or ''
        config_dict['harbor_db_sslmode'] = 'disable'
        config_dict['harbor_db_max_idle_conns'] = db_configs.get("max_idle_conns") or default_db_max_idle_conns
        config_dict['harbor_db_max_open_conns'] = db_configs.get("max_open_conns") or default_db_max_open_conns

        if with_clair:
            # clair db
            config_dict['clair_db_host'] = 'postgresql'
            config_dict['clair_db_port'] = 5432
            config_dict['clair_db_name'] = 'postgres'
            config_dict['clair_db_username'] = 'postgres'
            config_dict['clair_db_password'] = db_configs.get("password") or ''
            config_dict['clair_db_sslmode'] = 'disable'

        if with_notary:
            # notary signer
            config_dict['notary_signer_db_host'] = 'postgresql'
            config_dict['notary_signer_db_port'] = 5432
            config_dict['notary_signer_db_name'] = 'notarysigner'
            config_dict['notary_signer_db_username'] = 'signer'
            config_dict['notary_signer_db_password'] = 'password'
            config_dict['notary_signer_db_sslmode'] = 'disable'
            # notary server
            config_dict['notary_server_db_host'] = 'postgresql'
            config_dict['notary_server_db_port'] = 5432
            config_dict['notary_server_db_name'] = 'notaryserver'
            config_dict['notary_server_db_username'] = 'server'
            config_dict['notary_server_db_password'] = 'password'
            config_dict['notary_server_db_sslmode'] = 'disable'


    # Data path volume
    config_dict['data_volume'] = configs['data_volume']

    # Initial Admin Password
    config_dict['harbor_admin_password'] = configs["harbor_admin_password"]

    # Registry storage configs
    storage_config = configs.get('storage_service') or {}

    config_dict['registry_custom_ca_bundle_path'] = storage_config.get('ca_bundle') or ''

    if storage_config.get('filesystem'):
        config_dict['storage_provider_name'] = 'filesystem'
        config_dict['storage_provider_config'] = storage_config['filesystem']
    elif storage_config.get('azure'):
        config_dict['storage_provider_name'] = 'azure'
        config_dict['storage_provider_config'] = storage_config['azure']
    elif storage_config.get('gcs'):
        config_dict['storage_provider_name'] = 'gcs'
        config_dict['storage_provider_config'] = storage_config['gcs']
    elif storage_config.get('s3'):
        config_dict['storage_provider_name'] = 's3'
        config_dict['storage_provider_config'] = storage_config['s3']
    elif storage_config.get('swift'):
        config_dict['storage_provider_name'] = 'swift'
        config_dict['storage_provider_config'] = storage_config['swift']
    elif storage_config.get('oss'):
        config_dict['storage_provider_name'] = 'oss'
        config_dict['storage_provider_config'] = storage_config['oss']
    else:
        config_dict['storage_provider_name'] = 'filesystem'
        config_dict['storage_provider_config'] = {}

    if storage_config.get('redirect'):
        config_dict['storage_redirect_disabled'] = storage_config['redirect']['disabled']

    # Global proxy configs
    proxy_config = configs.get('proxy') or {}
    proxy_components = proxy_config.get('components') or []
    no_proxy_config = proxy_config.get('no_proxy')
    all_no_proxy = INTERNAL_NO_PROXY_DN
    if no_proxy_config:
        all_no_proxy |= set(no_proxy_config.split(','))

    for proxy_component in proxy_components:
      config_dict[proxy_component + '_http_proxy'] = proxy_config.get('http_proxy') or ''
      config_dict[proxy_component + '_https_proxy'] = proxy_config.get('https_proxy') or ''
      config_dict[proxy_component + '_no_proxy'] = ','.join(all_no_proxy)

    # Clair configs, optional
    clair_configs = configs.get("clair") or {}
    config_dict['clair_db'] = 'postgres'
    updaters_interval = clair_configs.get("updaters_interval", None)
    config_dict['clair_updaters_interval'] = 12 if updaters_interval is None else updaters_interval

    # Chart configs
    chart_configs = configs.get("chart") or {}
    config_dict['chart_absolute_url'] = chart_configs.get('absolute_url') or ''

    # jobservice config
    js_config = configs.get('jobservice') or {}
    config_dict['max_job_workers'] = js_config["max_job_workers"]
    config_dict['jobservice_secret'] = generate_random_string(16)

    # notification config
    notification_config = configs.get('notification') or {}
    config_dict['notification_webhook_job_max_retry'] = notification_config["webhook_job_max_retry"]

    # Log configs
    allowed_levels = ['debug', 'info', 'warning', 'error', 'fatal']
    log_configs = configs.get('log') or {}

    log_level = log_configs['level']
    if log_level not in allowed_levels:
        raise Exception('log level must be one of debug, info, warning, error, fatal')
    config_dict['log_level'] = log_level.lower()

    # parse local log related configs
    local_logs = log_configs.get('local') or {}
    if local_logs:
        config_dict['log_location'] = local_logs.get('location') or '/var/log/harbor'
        config_dict['log_rotate_count'] = local_logs.get('rotate_count') or 50
        config_dict['log_rotate_size'] = local_logs.get('rotate_size') or '200M'

    # parse external log endpoint related configs
    if log_configs.get('external_endpoint'):
        config_dict['log_external'] = True
        config_dict['log_ep_protocol'] = log_configs['external_endpoint']['protocol']
        config_dict['log_ep_host'] = log_configs['external_endpoint']['host']
        config_dict['log_ep_port'] = log_configs['external_endpoint']['port']
    else:
        config_dict['log_external'] = False

    # external DB, optional, if external_db enabled, it will cover the database config
    external_db_configs = configs.get('external_database') or {}
    if external_db_configs:
        config_dict['external_database'] = True
        # harbor db
        config_dict['harbor_db_host'] = external_db_configs['harbor']['host']
        config_dict['harbor_db_port'] = external_db_configs['harbor']['port']
        config_dict['harbor_db_name'] = external_db_configs['harbor']['db_name']
        config_dict['harbor_db_username'] = external_db_configs['harbor']['username']
        config_dict['harbor_db_password'] = external_db_configs['harbor']['password']
        config_dict['harbor_db_sslmode'] = external_db_configs['harbor']['ssl_mode']
        config_dict['harbor_db_max_idle_conns'] = external_db_configs['harbor'].get("max_idle_conns") or default_db_max_idle_conns
        config_dict['harbor_db_max_open_conns'] = external_db_configs['harbor'].get("max_open_conns") or default_db_max_open_conns

        if with_clair:
            # clair db
            config_dict['clair_db_host'] = external_db_configs['clair']['host']
            config_dict['clair_db_port'] = external_db_configs['clair']['port']
            config_dict['clair_db_name'] = external_db_configs['clair']['db_name']
            config_dict['clair_db_username'] = external_db_configs['clair']['username']
            config_dict['clair_db_password'] = external_db_configs['clair']['password']
            config_dict['clair_db_sslmode'] = external_db_configs['clair']['ssl_mode']
        if with_notary:
            # notary signer
            config_dict['notary_signer_db_host'] = external_db_configs['notary_signer']['host']
            config_dict['notary_signer_db_port'] = external_db_configs['notary_signer']['port']
            config_dict['notary_signer_db_name'] = external_db_configs['notary_signer']['db_name']
            config_dict['notary_signer_db_username'] = external_db_configs['notary_signer']['username']
            config_dict['notary_signer_db_password'] = external_db_configs['notary_signer']['password']
            config_dict['notary_signer_db_sslmode'] = external_db_configs['notary_signer']['ssl_mode']
            # notary server
            config_dict['notary_server_db_host'] = external_db_configs['notary_server']['host']
            config_dict['notary_server_db_port'] = external_db_configs['notary_server']['port']
            config_dict['notary_server_db_name'] = external_db_configs['notary_server']['db_name']
            config_dict['notary_server_db_username'] = external_db_configs['notary_server']['username']
            config_dict['notary_server_db_password'] = external_db_configs['notary_server']['password']
            config_dict['notary_server_db_sslmode'] = external_db_configs['notary_server']['ssl_mode']
    else:
        config_dict['external_database'] = False

    # update redis configs
    config_dict.update(get_redis_configs(configs.get("external_redis", None), with_clair))

    # auto generated secret string for core
    config_dict['core_secret'] = generate_random_string(16)

     # Admiral configs
    config_dict['admiral_url'] = configs.get("admiral_url") or ""

    # UAA configs
    config_dict['uaa'] = configs.get('uaa') or {}

    return config_dict


def get_redis_url(db, redis=None):
    """Returns redis url with format `redis://[arbitrary_username:password@]ipaddress:port/database_index`

    >>> get_redis_url(1)
    'redis://redis:6379/1'
    >>> get_redis_url(1, {'host': 'localhost', 'password': 'password'})
    'redis://anonymous:password@localhost:6379/1'
    """
    kwargs = {
        'host': 'redis',
        'port': 6379,
        'password': '',
    }
    kwargs.update(redis or {})
    kwargs['db'] = db

    if kwargs['password']:
        return "redis://anonymous:{password}@{host}:{port}/{db}".format(**kwargs)
    return "redis://{host}:{port}/{db}".format(**kwargs)


def get_redis_configs(external_redis=None, with_clair=True):
    """Returns configs for redis

    >>> get_redis_configs()['external_redis']
    False
    >>> get_redis_configs()['redis_url_reg']
    'redis://redis:6379/1'
    >>> get_redis_configs()['redis_url_js']
    'redis://redis:6379/2'
    >>> get_redis_configs()['redis_url_clair']
    'redis://redis:6379/4'

    >>> get_redis_configs({'host': 'localhost', 'password': ''})['redis_password']
    ''
    >>> get_redis_configs({'host': 'localhost', 'password': None})['redis_password']
    ''
    >>> get_redis_configs({'host': 'localhost', 'password': None})['redis_url_reg']
    'redis://localhost:6379/1'

    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['external_redis']
    True
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_password']
    'pass'
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_url_reg']
    'redis://anonymous:pass@localhost:6379/1'
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_url_js']
    'redis://anonymous:pass@localhost:6379/2'
    >>> get_redis_configs({'host': 'localhost', 'password': 'pass'})['redis_url_clair']
    'redis://anonymous:pass@localhost:6379/4'

    >>> 'redis_url_clair' not in get_redis_configs(with_clair=False)
    True
    """
    external_redis = external_redis or {}

    configs = dict(external_redis=bool(external_redis))

    # internal redis config as the default
    redis = {
        'host': 'redis',
        'port': 6379,
        'password': '',
        'registry_db_index': 1,
        'jobservice_db_index': 2,
        'chartmuseum_db_index': 3,
        'clair_db_index': 4,
    }

    # overwriting existing keys by external_redis
    redis.update({key: value for (key, value) in external_redis.items() if value})

    configs['redis_host'] = redis['host']
    configs['redis_port'] = redis['port']
    configs['redis_password'] = redis['password']
    configs['redis_db_index_reg'] = redis['registry_db_index']
    configs['redis_db_index_js'] = redis['jobservice_db_index']
    configs['redis_db_index_chart'] = redis['chartmuseum_db_index']

    configs['redis_url_js'] = get_redis_url(configs['redis_db_index_js'], redis)
    configs['redis_url_reg'] = get_redis_url(configs['redis_db_index_reg'], redis)

    if with_clair:
        configs['redis_db_index_clair'] = redis['clair_db_index']
        configs['redis_url_clair'] = get_redis_url(configs['redis_db_index_clair'], redis)

    return configs
