# go run main.go -generate
# ENV_FILE=".production.env"
#
# CURRENT_DATE=$(date +"%b %e, %Y")
#
# if grep -q "^UPDATED_AT=" "$ENV_FILE"; then
#     sed -i "s/^UPDATED_AT=.*/UPDATED_AT=\"$CURRENT_DATE\"/" "$ENV_FILE"
# else
#     echo "UPDATED_AT=$CURRENT_DATE" >> "$ENV_FILE"
# fi
#
# bun tailwindcss -i ./input.css -o ./static/css/style.css
# templ generate
#
# # ./generate_resume_pdf.sh
#
# # make sure that hitcounters.db is there
# ssh root@404notboring.com "touch /root/404notboring.com/hitcounters.db"
#
# docker build -t 404notboring.com:latest .
# docker save -o 404notboring.com.tar 404notboring.com:latest
# rsync -uvrP 404notboring.com.tar root@404notboring.com:/root/404notboring.com/
# rsync -uvrP docker-compose.yaml root@404notboring.com:/root/404notboring.com/
# ssh root@404notboring.com "docker load -i /root/404notboring.com/404notboring.com.tar"
# ssh root@404notboring.com "cd /root/404notboring.com && docker compose up -d && docker image prune -f"
