select * from annotations where hash_id = :hash_id;
select * from properties where hash_id = :hash_id;
select * from tags where hash_id = :hash_id;
select encode(bytes, 'escape') from contents where hash_id = :hash_id;

