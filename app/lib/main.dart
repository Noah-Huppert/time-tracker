import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'dart:async';

final GoogleSignIn _googleSignIn = GoogleSignIn();
final FirebaseAuth _auth = FirebaseAuth.instance;

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Time Tracker',
      home: Scaffold(
        appBar: AppBar(
          title: Text('Time Tracker'),
        ),
        body: Center(
          child: Container(
            child: TimerWidget(),
          ),
        ),
      ),
    );
  }
}

class TimerWidgetState extends State<TimerWidget> {
  Timer timer;
  
  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        Center(
          child: _buildTopWidget(),
        ),
        Row(
          children: <Widget>[
            RaisedButton(
              onPressed: () {
                // Start timer if not running
                if (timer == null || !timer.isActive) {
                  setState(() {
                    timer = Timer.periodic(Duration(seconds: 1), (timer){});
                  });
                } else { // Stop timer if running
                  setState(() {
                      timer.cancel();
                  });
                }
              },
              child: _buildToggleWidget(),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildTopWidget() {
    return Text("TOP WIDGET");
  }

  Widget _buildToggleWidget() {
    if (timer != null && timer.isActive) {
      return Text("Stop");
    } else {
      return Text("Start");
    }
  }
}

class TimerWidget extends StatefulWidget {
  @override
  TimerWidgetState createState() => TimerWidgetState();
}
