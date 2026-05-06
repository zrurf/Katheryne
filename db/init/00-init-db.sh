#!/bin/bash
set -e

# 获取数据库环境变量，默认值为空
DATABASES=${DATABASES:-}
if [ -z "$DATABASES" ]; then
    echo "WARNING: Environment variable DATABASES not set. No databases will be created."
    exit 0
fi

DB_LIST=$(echo "$DATABASES" | tr ',' ' ')

# 基础 SQL 目录，预期结构：/etc/sql/<db_name>/init/*.sql
BASE_SQL_DIR="/etc/sql"

PG_OPTIONS="--username $POSTGRES_USER --dbname postgres"

for db in $DB_LIST; do
    echo "Initializing database: $db"

    # 检查数据库是否存在，不存在则创建
    if psql $PG_OPTIONS -tAc "SELECT 1 FROM pg_database WHERE datname='$db'" | grep -q 1; then
        echo "Database '$db' already exists, skipping creation."
    else
        echo "Creating database '$db'..."
        psql $PG_OPTIONS -c "CREATE DATABASE $db"

        # 授权用户
        if [ -n "$POSTGRES_USER" ]; then
            psql $PG_OPTIONS -c "GRANT ALL PRIVILEGES ON DATABASE $db TO $POSTGRES_USER"
        fi
    fi

    # 执行该数据库专属 init 目录下的所有 .sql 文件
    DB_SQL_DIR="$BASE_SQL_DIR/$db/init"
    if [ -d "$DB_SQL_DIR" ]; then
        echo "Found SQL directory for $db: $DB_SQL_DIR"
        # 按文件名排序执行
        for sql_file in $(ls -v $DB_SQL_DIR/*.sql 2>/dev/null || true); do
            if [ -f "$sql_file" ]; then
                echo "Executing $sql_file on database $db..."
                psql --username $POSTGRES_USER --dbname $db -f "$sql_file"
            fi
        done
    else
        echo "Directory $DB_SQL_DIR does not exist or is empty. No SQL files executed for $db."
    fi
done

echo "Initialization complete."