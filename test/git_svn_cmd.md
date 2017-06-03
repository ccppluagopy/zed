svn checkout remote_url local_dir;
svn update -r r10037;
svn update;

git clone remote_url local_dir;
git pull  remote_url;
git checkout v_sha1;
