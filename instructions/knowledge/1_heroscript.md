# HeroScript

## Overview

HeroScript is a simple, declarative scripting language designed to define workflows and execute commands in a structured manner. It follows a straightforward syntax where each action is prefixed with `!!`, indicating the actor and action name.

## Example

A basic HeroScript script for virtual machine management looks like this:

```heroscript
!!vm.define name:'test_vm' cpu:4
    memory: '8GB'
    storage: '100GB'
	description: '
		A virtual machine configuration
		with specific resources.
	'

!!vm.start name:'test_vm'

!!vm.disk_add
	name: 'test_vm'
	size: '50GB'
	type: 'SSD'

!!vm.delete
	name: 'test_vm'
	force: true
```

### Key Features

- Every action starts with `!!`.
  - The first part after `!!` is the actor (e.g., `vm`).
  - The second part is the action name (e.g., `define`, `start`, `delete`).
- Multi-line values are supported (e.g., the `description` field).
- Lists are comma-separated where applicable and inside ''.
- If items one 1 line, then no space between name & argument e.g. name:'test_vm'

## Parsing HeroScript

Internally, HeroScript gets parsed into an action object with parameters. Each parameter follows a `key: value` format.

### Parsing Example

```heroscript
!!actor.action
    id:a1 name6:aaaaa
    name:'need to do something 1' 
    description:
        '
        ## markdown works in it
        description can be multiline
        lets see what happens

        - a
        - something else

        ### subtitle
        '

    name2:   test
    name3: hi 
    name10:'this is with space'  name11:aaa11

    name4: 'aaa'

    //somecomment
    name5:   'aab'
```

### Parsing Details
- Each parameter follows a `key: value` format.
- Multi-line values (such as descriptions) support Markdown formatting.
- Comments can be added using `//`.
- Keys and values can have spaces, and values can be enclosed in single quotes.

