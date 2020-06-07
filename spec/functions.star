def get_password(users, user):
    return {x.username: x.password for x in users}[user]
end

def csv(users):
  return '\n'.join(['username,password'] +
         [x.username + ',' + x.password for x in list(users)]) 
end
