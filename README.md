# What is this?

Hakscale allows you to scale out mass-commands over multiple systems, with multiple threads on each system.

# How does it work?

Hakscale can run in two modes. Push, and pop.

In push mode, hakscale will push specified commands to a redis queue, wait for output to come back, then print it.
In pop mode, hakscale will pop commands off the redis queue, execute the command, then send back the output.