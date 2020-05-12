# GoCostumerSupportApi

Updates for the customer ticketing system

Now we're adding real customer support people to answer the questions, this would involve changes in the API,
and of course the code changes too, you will need to drop certain logic, so better create a copy of the initial
project and do your changes there, so that you can keep the previous version separately.

API changes:

1. POST /api/question/post method:
Note that the URL has changed - now we have "/api/question/post".

Now that there will be real people answering the questions, their response will take some time,
so now the /api/question/post method will return the id of the ticket, so that later the person who asked
the question can use this id to check the status of the question.

Request is still the same:
    {
      "priority": <priority>,
      "question": <text of the question>
    }
    But response will now be:
    {
      "success": <true|false>,
      "id": <created ticket id or empty on error>,
      "message": <"OK" on success|error if the request format is incorrect>
    }

2. GET /api/status - some changes in the response:
    {
      "questions_processed": <total number of answered questions>,
      "average_response_time": <average time for a request to be answered>,
      "queue_length":
        {
          1: <length of first priority queue>,
          2: <length of second priority queue>,
          1: <length of third priority queue>
        },
      "questions_submitted": <total number of questions submitted>,
      "questions_answered": <total number of questions answered>
    }

    Please note that we removed here the number of workers (as we won't have automated workers).

3. New method: GET /api/question/get_next
This is the method to be used by the customer support people, it's a simple GET
request without parameters, the API should return json response like this:
  {
    "id": <question id>,
    "priority": <question priority>,
    "question": <text of the question>
  }

4. New method: POST /api/question/answer
This method will be used by customer support people to answer the question, request format:
 {
   "id": <question id that is being answered>,
   "answer": <text of the answer>
 }

 Response should be in the form:
 {
   "success": <true|false>,
   "message": <error if the request has wrong format or question id not found>
 }

5. New method: GET /api/question/status
This method will be used by the customers - to check if there are any updates
regarding their question.
Request query string: ?question_id=<question_id>, so for example, the full request
URL will look like this: /api/question/status?question_id=1

Response will look like this:
  {
    "id": <question_id>,
    "question": <text of the question>,
    "status": <queued|in_progress|answered>,  
    "answer": <answer if it's ready, otherwise empty string>,
    "answer_time": <time when the question was answered, or empty otherwise"
  }

Note that the status field can have the following values:
 - "queued" means that the question was added to the queue, but not yet picked up
by customer support (/api/question/get_next has not returned this question to
the customer support person yet).
 - "in_progress" means that the customer support person retrieved the question,
 but the answer has not been submitted yet.
 - "answered" means that the answer to the question was submitted.



