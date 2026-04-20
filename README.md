# Purchases
Жена попросила создать программу для покупок, посмотрим что из этого выйдет


Описание: Программа должна хранить список покупок, чтобы можно было заполнять на одном устройстве, и читать на другом. Также там должна быть история, и какая-никая статистика.

Для начала нужно сделать базу:
users:
id, name, password, email, role, status, created_at, update_at

products:
id, title, group, unit

store:
id, name

buy:
id, users_id, store_id, products_id, count, added_at

history:
id, user_id, store_id, products_id, count, price, date
