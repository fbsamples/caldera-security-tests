# Findings

## Stored XSS #1

- Description: The first stored XSS was found during DEF CON 30 and was reported
  to the team at that time.

- Manual Reproduction:

0. Create a Caldera test environment
1. Login to Caldera with the red user
2. Click **operations** in the left-hand menu
3. Click **Create Operation**
4. Provide the following input for the name:

   ```js
   "><img src=x onerror=prompt(1)>
   ```

5. Click **Start**
6. Observe prompt:
   ![xss1](images/xss1.png)

- Remediation:

TODO: DEFENDERS

- State: [Remediated](https://github.com/mitre/caldera/pull/2644)

## Stored XSS #2

- Description: The second stored XSS was found after DEF CON 30.

- Manual Reproduction:

0. Create a Caldera test environment
1. Login to Caldera with the red user
2. Click **operations** in the left-hand menu
3. Click **Create Operation**
4. Provide the following input for the name:

   ```js
   "><img src=x onerror=prompt(2)>
   ```

5. Click **Start**
6. Click **debrief** in the left-hand menu
7. Select the operation created in the previous steps
8. Click the dropdown menu that
   currently reads **Attack Path** and change it to **Tactic**
9. Move your cursor over the name of the operation
10. Observe prompt:
    ![xss2](images/xss2.png)

- Remediation:

TODO: DEFENDERS

- State: **Vulnerable**
