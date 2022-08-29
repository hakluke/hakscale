# What is this?

Hakscale allows you to scale out shell commands over multiple systems with multiple threads on each system. The key concept is that a master server will _push_ commands to the queue, then multiple worker servers _pop_ commands from the queue and execute them.

For example, if you want to run a tool like httpx against 1 million hosts in `hosts.txt`, you could run:

```
hakscale push -p "host:./hosts.txt" -c "echo _host_ | httpx" -t 20
```

This would create 1 million commands and send them to a queue. Then, you can set up as many servers (workers) as you like to pull those commands off the queue and execute them. To do this, you can simply run the following command:

```
hakscale pop -t 20
```

Once the command is complete, the output is sent back to the master server that pushed the commands in the first place.

# What can it be used for?

It's a very simple way to distribute scans/commands across many systems. It's perfect for large-scale internet scanning because it is _way_ faster than attempting to do it from a single host. There are probably also a bunch of other uses that I can't think of right now.

# Setup

The basic requirements are:

- A computer to push commands
- 1 or more worker computers (usually multiple VPSs)
- A Redis server

You can set it up however you want, but if you would like to use Digital Ocean:

- Set up Redis server ([Digital Ocean can do this for you in seconds, click here for $100 credit](https://m.do.co/c/ac22891d18e8))
  - Log in to Digital Ocean
  - Create > Databases > Redis
  - Follow the prompts to create the database and save the details
- Spin up a bunch of droplets (or any VPS) to use as workers
  - Create > Droplets
- Set up hakscale config file to include the Redis server details
  - Save the following to `~/.config/haktools/hakscale-config.yml` on every VPS/computer that you want to use hakscale with

```
redis:
  host: your-redis-host
  port: your-redis-port
  password: your-redis-password
```

- Run `hakscale pop` on the workers

You're now ready to start distributing your commands!
