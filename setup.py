from setuptools import find_packages, setup

setup(name='syncmeister',
      version='0.0.1',
      description='Fitness data syncing tool',
      author='Brian Elliott',
      author_email='bdelliott@protonmail.com',
      license='MIT',
      install_requires=['requests'],
      namespace_packages=[],
      packages=find_packages('.'),
      include_package_data=True,
      classifiers=[
          'Development Status :: 3 - Alpha',
          'Environment :: Console',
          'Intended Audience :: Developers',
          'Programming Language :: Python :: 3.6',
      ],
      entry_points={
        'console_scripts': ['nokia=syncmeister.nokia:cli']  
      }
)
