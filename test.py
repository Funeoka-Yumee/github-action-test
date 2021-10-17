import os
import random
import string
import time

os.system('dig a6008.com A @120.79.177.7')

for i in range(2500):
    random_str = ''.join([random.choice(string.ascii_letters+string.digits) for i in range(6)])
    domain = '{}.a6008.com'.format(random_str)
    os.system(f'dig {domain} A @120.79.177.7')
    time.sleep(0.05)
