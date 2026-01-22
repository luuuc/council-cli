# Remove Expert from Council

Remove an expert from the council: $ARGUMENTS

## Instructions

You are removing an expert from the council.

## Step 1: Identify the Expert

Parse the arguments to get the expert name or ID. If not provided, list current experts:

```bash
council list
```

Then ask the user which expert to remove.

## Step 2: Remove the Expert

Run the council remove command with the expert ID:

```bash
council remove {expert-id}
```

The command will ask for confirmation before removing.

## Step 3: Sync Changes

After removal, sync the changes to AI tool configurations:

```bash
council sync
```

## After Removing

Confirm with: "Removed {Name} from the council"
